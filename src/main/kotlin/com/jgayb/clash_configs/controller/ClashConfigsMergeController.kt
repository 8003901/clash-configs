package com.jgayb.clash_configs.controller

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.node.ArrayNode
import com.fasterxml.jackson.databind.node.ObjectNode
import com.fasterxml.jackson.dataformat.yaml.YAMLMapper
import com.jgayb.clash_configs.eneity.ClashConfig
import com.jgayb.clash_configs.eneity.ClashConfigsMerge
import com.jgayb.clash_configs.service.ClashConfigsMergeService
import jakarta.validation.constraints.NotBlank
import org.springframework.http.HttpStatus
import org.springframework.validation.annotation.Validated
import org.springframework.web.bind.annotation.*
import org.springframework.web.server.ResponseStatusException

@RestController
@RequestMapping("/clash_configs_merge")
class ClashConfigsMergeController(
    val clashConfigsMergeService: ClashConfigsMergeService
) {
    @PostMapping
    fun create(@RequestBody @Validated clashConfigsMergeAdd: ClashConfigsMergeAdd): ClashConfigsMerge {
        return clashConfigsMergeService.create(clashConfigsMergeAdd.name, clashConfigsMergeAdd.configIds)
    }

    @PutMapping
    fun update(@Validated @RequestBody clashConfigsMergeUpdate: ClashConfigsMergeUpdate): ClashConfigsMerge {
        return ClashConfigsMerge()
    }

    @PutMapping("/{id}/token")
    fun refreshToken(@PathVariable id: String): ClashConfigsMerge {
        return clashConfigsMergeService.refreshToken(id)
    }


}

@RestController
@RequestMapping("/configs")
class Config(
    private val clashConfigsMergeService: ClashConfigsMergeService, private val objectMapper: ObjectMapper
) {
    private var yamlMapper: YAMLMapper = YAMLMapper()

    @GetMapping(produces = ["text/plain"])
    fun query(@RequestParam token: String): String {
        val configsMerge = clashConfigsMergeService.findByToken(token) ?: throw ResponseStatusException(
            HttpStatus.NOT_FOUND,
            "Config from token [$token] not found"
        )
        val tempNode = objectMapper.readTree(configsMerge.config)
        val proxies = tempNode.get("proxies") as ArrayNode
        val pxArrayNode = objectMapper.createArrayNode()
        configsMerge.configs?.filter {
            it.content?.isNotEmpty() == true
        }?.forEach {
            val pn = yamlMapper.readTree(it.content).get("proxies")
            if (pn?.isArray == true) {
                proxies.addAll(pn as ArrayNode)
            }
        }
        proxies.map {
            it["name"].asText()
        }.forEach(pxArrayNode::add)
        tempNode.get("proxy-groups").forEach {
            val filterKey = it.get("filter-key").asText("")
            val od = it as ObjectNode
            if (filterKey == "all") {
                val op = od.get("proxies")
                if (op != null && op.isArray) {
                    (op as ArrayNode).addAll(pxArrayNode)
                } else {
                    od.put("proxies", pxArrayNode)
                }
            } else if (filterKey.isNotEmpty()) {
                val keys = filterKey.split("|")
                val gpt = objectMapper.createArrayNode()
                proxies.filter { p ->
                    val n = p["name"].asText()
                    keys.any { k ->
                        n.contains(k)
                    }
                }.map { p ->
                    p["name"].asText()
                }.forEach(gpt::add)
                val op = od.get("proxies")
                if (op != null && op.isArray) {
                    (op as ArrayNode).addAll(gpt)
                } else {
                    od.put("proxies", gpt)
                }
            }
            od.remove("filter-key")
        }
        return yamlMapper.writeValueAsString(tempNode).replaceFirst("---\n", "")
    }

}

data class ClashConfigsMergeAdd(
    @NotBlank var name: String, var configIds: MutableList<String>? = null
)

data class ClashConfigsMergeUpdate(
    @NotBlank var id: String? = null,
    @NotBlank var name: String? = null,
    @NotBlank var config: String? = null,
    var configs: MutableList<ClashConfig>? = null
)