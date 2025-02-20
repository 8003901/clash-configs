package com.jgayb.clash_configs.controller

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.node.ArrayNode
import com.fasterxml.jackson.databind.node.ObjectNode
import com.fasterxml.jackson.dataformat.yaml.YAMLMapper
import com.jgayb.clash_configs.eneity.ClashConfig
import com.jgayb.clash_configs.eneity.ClashConfigsMerge
import com.jgayb.clash_configs.service.ClashConfigsMergeService
import jakarta.validation.constraints.NotBlank
import org.springframework.beans.BeanUtils
import org.springframework.http.HttpStatus
import org.springframework.validation.annotation.Validated
import org.springframework.web.bind.annotation.*
import org.springframework.web.server.ResponseStatusException

@RestController
@RequestMapping("/clash_configs_merge")
class ClashConfigsMergeController(
    private val clashConfigsMergeService: ClashConfigsMergeService
) {
    @PostMapping
    fun create(@RequestBody @Validated clashConfigsMergeAdd: ClashConfigsMergeAdd): ClashConfigsMerge {
        return clashConfigsMergeService.create(clashConfigsMergeAdd.name, clashConfigsMergeAdd.configIds)
    }

    @PutMapping
    fun update(@Validated @RequestBody clashConfigsMergeUpdate: ClashConfigsMergeUpdate): ClashConfigsMerge {
        val ccm = ClashConfigsMerge()
        BeanUtils.copyProperties(clashConfigsMergeUpdate, ccm)
        return clashConfigsMergeService.save(ccm)
    }

    @PutMapping("/{id}/token")
    fun refreshToken(@PathVariable id: String): ClashConfigsMerge {
        return clashConfigsMergeService.refreshToken(id)
    }
}

@RestController
@RequestMapping("/configs")
class ConfigController(
    private val clashConfigsMergeService: ClashConfigsMergeService,
    private val objectMapper: ObjectMapper
) {
    private val yamlMapper = YAMLMapper()

    @GetMapping(produces = ["text/plain"])
    fun query(@RequestParam token: String): String {
        val configsMerge = clashConfigsMergeService.findByToken(token) ?: throw ResponseStatusException(
            HttpStatus.NOT_FOUND,
            "Config from token [$token] not found"
        )
        return processConfig(configsMerge)
    }

    private fun processConfig(configsMerge: ClashConfigsMerge): String {
        val tempNode = objectMapper.readTree(configsMerge.config)
        val proxies = tempNode.get("proxies") as ArrayNode
        val proxyNames = objectMapper.createArrayNode()

        configsMerge.configs
            ?.filter { it.content?.isNotEmpty() == true }
            ?.forEach { config ->
                yamlMapper.readTree(config.content)
                    .get("proxies")
                    ?.takeIf { it.isArray }
                    ?.let { proxies.addAll(it as ArrayNode) }
            }

        proxies.mapNotNull { it["name"]?.asText() }
            .forEach(proxyNames::add)

        tempNode.get("proxy-groups").forEach { group ->
            processProxyGroup(group as ObjectNode, proxyNames, proxies)
        }

        return yamlMapper.writeValueAsString(tempNode).replaceFirst("---\n", "")
    }

    private fun processProxyGroup(group: ObjectNode, proxyNames: ArrayNode, proxies: ArrayNode) {
        val filterKey = group.get("filter-key")?.asText("") ?: ""

        when {
            filterKey == "all" -> updateProxies(group, proxyNames)
            filterKey.isNotEmpty() -> {
                val keys = filterKey.split("|")
                val filteredProxies = objectMapper.createArrayNode()

                proxies.filter { proxy ->
                    val name = proxy["name"].asText()
                    keys.any { key -> name.contains(key) }
                }.forEach { proxy ->
                    filteredProxies.add(proxy["name"].asText())
                }

                updateProxies(group, filteredProxies)
            }
        }

        group.remove("filter-key")
    }

    private fun updateProxies(group: ObjectNode, newProxies: ArrayNode) {
        val existingProxies = group.get("proxies")
        if (existingProxies?.isArray == true) {
            (existingProxies as ArrayNode).addAll(newProxies)
        } else {
            group.put("proxies", newProxies)
        }
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