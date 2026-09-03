package com.jgayb.clash_configs.service

import com.jgayb.clash_configs.eneity.ClashConfig

interface ClashConfigService {

    fun saveClashConfig(clashConfig: ClashConfig): ClashConfig

    fun detail(id: String, renew: Boolean): ClashConfig

    fun allClashConfigs(): List<ClashConfig>

    fun deleteClashConfig(id: String)

}