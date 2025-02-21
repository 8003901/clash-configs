package com.jgayb.clash_configs.controller

import com.jgayb.clash_configs.eneity.ClashConfig
import com.jgayb.clash_configs.eneity.UpdateSchedule
import com.jgayb.clash_configs.service.ClashConfigService
import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.NotNull
import org.springframework.validation.annotation.Validated
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController


@RestController
@RequestMapping("/clash_configs")
class ClashConfigController(
    private val clashConfigService: ClashConfigService,
) {


    @PostMapping
    fun saveOrUpdateClashConfig(@RequestBody @Validated clashConfigAdd: ClashConfigAdd): ClashConfig {
        val cc = ClashConfig()
        cc.id = clashConfigAdd.id
        cc.name = clashConfigAdd.name
        cc.url = clashConfigAdd.url
        cc.updateSchedule = clashConfigAdd.updateSchedule
        cc.enabled = clashConfigAdd.enabled
        return clashConfigService.saveClashConfig(cc)
    }

    @GetMapping
    fun list(): List<ClashConfig> {
        return clashConfigService.allClashConfigs()
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