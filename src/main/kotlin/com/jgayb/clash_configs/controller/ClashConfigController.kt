package com.jgayb.clash_configs.controller

import com.jgayb.clash_configs.eneity.ClashConfig
import com.jgayb.clash_configs.eneity.UpdateSchedule
import com.jgayb.clash_configs.service.ClashConfigService
import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.NotNull
import org.springframework.http.HttpEntity
import org.springframework.http.HttpHeaders
import org.springframework.http.HttpMethod
import org.springframework.validation.annotation.Validated
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController
import org.springframework.web.client.RestTemplate


@RestController
@RequestMapping("/clash_configs")
class ClashConfigController(val clashConfigService: ClashConfigService) {

    val restTemplate: RestTemplate = RestTemplate()

    @PostMapping
    fun saveOrUpdateClashConfig(@RequestBody @Validated clashConfigAdd: ClashConfigAdd): ClashConfig {
        val cc = ClashConfig()
        cc.id = clashConfigAdd.id
        cc.name = clashConfigAdd.name
        cc.url = clashConfigAdd.url
        cc.updateSchedule = clashConfigAdd.updateSchedule
        cc.enabled = clashConfigAdd.enabled
        val headers = HttpHeaders()
        headers.add("User-Agent", "clash-verge/*")
        // 创建 HttpEntity，包含请求头
        val entity = HttpEntity<String>(headers)
        restTemplate.exchange(
            clashConfigAdd.url!!, HttpMethod.GET,
            entity,
            String::class.java
        ).body?.let {
            cc.content = it
        }

//        restTemplate.getForObject(clashConfigAdd.url!!, String::class.java)?.let {
//            cc.content = String(Base64.getDecoder().decode(it))
//        }
        return clashConfigService.saveClashConfig(cc)
    }


}

data class ClashConfigAdd(
    var id: String? = null,
    @NotBlank
    var url: String? = null,
    @NotBlank
    var name: String? = null,
    @NotNull
    var updateSchedule: UpdateSchedule = UpdateSchedule.DAY,
    var enabled: Boolean = false,
)