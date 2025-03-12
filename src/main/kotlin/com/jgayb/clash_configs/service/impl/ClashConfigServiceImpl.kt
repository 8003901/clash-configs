package com.jgayb.clash_configs.service.impl

import com.jgayb.clash_configs.eneity.ClashConfig
import com.jgayb.clash_configs.eneity.UpdateSchedule
import com.jgayb.clash_configs.repository.ClashConfigRepository
import com.jgayb.clash_configs.service.ClashConfigService
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.beans.BeanUtils
import org.springframework.data.domain.Example
import org.springframework.http.HttpEntity
import org.springframework.http.HttpHeaders
import org.springframework.http.HttpMethod
import org.springframework.http.HttpStatus
import org.springframework.scheduling.annotation.Scheduled
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import org.springframework.web.client.RestTemplate
import org.springframework.web.server.ResponseStatusException
import java.util.*

@Service
class ClashConfigServiceImpl(
    private val clashConfigRepository: ClashConfigRepository,
    private val restTemplate: RestTemplate
) : ClashConfigService {
    private var log: Logger = LoggerFactory.getLogger(this.javaClass)

    @Scheduled(cron = "0 0 2 * * ?")
    @Transactional
    fun renewConfigDaily() {
        clashConfigRepository.streamAll(UpdateSchedule.DAY)
            ?.forEach { clashConfig ->
                try {
                    fetchConfigFromRemote(clashConfig)
                    log.info("每日更新 ${clashConfig.name}")
                    clashConfigRepository.save(clashConfig)
                } catch (ex: Exception) {
                    log.error("Error during clash config renew", ex)
                }
            }
    }

    @Scheduled(cron = "0 0 1 ? * MON")
    @Transactional
    fun renewConfigWeekly() {
        clashConfigRepository.streamAll(UpdateSchedule.WEEK)
            ?.forEach { clashConfig ->
                fetchConfigFromRemote(clashConfig)
                log.info("每周更新 ${clashConfig.name}")
                clashConfigRepository.save(clashConfig)
            }
    }

    private fun fetchConfigFromRemote(clashConfig: ClashConfig) {
        val headers = HttpHeaders()
        headers.add("User-Agent", "clash-verge/*")
        // 创建 HttpEntity，包含请求头
        val entity = HttpEntity<String>(headers)
        val responseEntity = restTemplate.exchange(
            clashConfig.url!!, HttpMethod.GET,
            entity,
            String::class.java
        )
        clashConfig.subscriptionUserinfo = responseEntity.headers.getFirst("subscription-userinfo")
        responseEntity.body?.let {
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
                it.url = clashConfig.url
                it.subscriptionUserinfo = clashConfig.subscriptionUserinfo
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

    override fun detail(id: String, renew: Boolean): ClashConfig {
        val opt = clashConfigRepository.findById(id)
        if (!opt.isPresent) {
            throw ResponseStatusException(HttpStatus.NOT_FOUND)
        }
        val config = opt.get()
        if (renew) {
            fetchConfigFromRemote(config)
            config.updatedAt = Date()
            return clashConfigRepository.save(config)
        }
        return config
    }

    override fun allClashConfigs(): List<ClashConfig> {
        val example = Example.of(ClashConfig(enabled = true))
        return clashConfigRepository.findAll(example)
    }
}