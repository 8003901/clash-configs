package com.jgayb.clash_configs.service.impl

import com.jgayb.clash_configs.eneity.ClashConfig
import com.jgayb.clash_configs.repository.ClashConfigRepository
import com.jgayb.clash_configs.service.ClashConfigService
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.beans.BeanUtils
import org.springframework.stereotype.Service
import java.util.*

@Service
class ClashConfigServiceImpl(val clashConfigRepository: ClashConfigRepository) : ClashConfigService {
    private var log: Logger = LoggerFactory.getLogger(this.javaClass)

    override fun saveClashConfig(clashConfig: ClashConfig): ClashConfig {
        if (!clashConfig.id.isNullOrEmpty()) {
            clashConfigRepository.findById(clashConfig.id!!).ifPresent {
                if (clashConfig.name?.isNotEmpty() == true) {
                    it.name = clashConfig.name
                }
                it.enabled = clashConfig.enabled
                it.content = clashConfig.content
                it.updatedAt = Date()
                clashConfigRepository.save(it)
                BeanUtils.copyProperties(it, clashConfig)
            }
            return clashConfig
        }
        clashConfig.id = UUID.randomUUID().toString()
        clashConfig.createdAt = Date()
        clashConfig.updatedAt = Date()
        return clashConfigRepository.save(clashConfig)
    }
}