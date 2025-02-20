package com.jgayb.clash_configs.service

import com.jgayb.clash_configs.eneity.ClashConfig

interface ClashConfigService {

    fun saveClashConfig(clashConfig: ClashConfig): ClashConfig

}