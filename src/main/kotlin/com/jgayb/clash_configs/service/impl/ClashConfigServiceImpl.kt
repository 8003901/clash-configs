package com.jgayb.clash_configs.service.impl

import com.jgayb.clash_configs.eneity.ClashConfig
import com.jgayb.clash_configs.eneity.UpdateSchedule
import com.jgayb.clash_configs.repository.ClashConfigRepository
import com.jgayb.clash_configs.service.ClashConfigService
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.beans.BeanUtils
import org.springframework.http.HttpEntity
import org.springframework.http.HttpHeaders
import org.springframework.http.HttpMethod
import org.springframework.scheduling.annotation.Scheduled
import org.springframework.stereotype.Service
import org.springframework.web.client.RestTemplate
import java.util.*

@Service
class ClashConfigServiceImpl(
    private val clashConfigRepository: ClashConfigRepository,
    private val restTemplate: RestTemplate
) : ClashConfigService {
    private var log: Logger = LoggerFactory.getLogger(this.javaClass)

    @Scheduled(cron = "0 0 2 * * ?")
    fun renewConfigDaily() {
        clashConfigRepository.streamAll(UpdateSchedule.DAY)
            ?.forEach { clashConfig ->
                fetchConfigFromRemote(clashConfig)
                clashConfigRepository.save(clashConfig)
            }
    }

    @Scheduled(cron = "0 0 1 ? * MON")
    fun renewConfigWeekly() {
        clashConfigRepository.streamAll(UpdateSchedule.WEEK)
            ?.forEach { clashConfig ->
                fetchConfigFromRemote(clashConfig)
                clashConfigRepository.save(clashConfig)
            }
    }

    private fun fetchConfigFromRemote(clashConfig: ClashConfig) {
        val headers = HttpHeaders()
        headers.add("User-Agent", "clash-verge/*")
        // 创建 HttpEntity，包含请求头
        val entity = HttpEntity<String>(headers)
        restTemplate.exchange(
            clashConfig.url!!, HttpMethod.GET,
            entity,
            String::class.java
        ).body?.let {
            clashConfig.content = it
        }
    }

    override fun saveClashConfig(clashConfig: ClashConfig): ClashConfig {
        fetchConfigFromRemote(clashConfig)
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