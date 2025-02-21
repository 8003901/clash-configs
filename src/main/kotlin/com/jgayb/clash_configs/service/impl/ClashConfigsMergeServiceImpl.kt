package com.jgayb.clash_configs.service.impl

import com.jgayb.clash_configs.eneity.ClashConfigsMerge
import com.jgayb.clash_configs.repository.ClashConfigRepository
import com.jgayb.clash_configs.repository.ClashConfigsMergeRepository
import com.jgayb.clash_configs.service.ClashConfigsMergeService
import jakarta.annotation.PostConstruct
import org.springframework.core.io.ClassPathResource
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import org.springframework.util.StreamUtils
import org.springframework.web.server.ResponseStatusException
import java.nio.charset.StandardCharsets
import java.util.*


@Service
open class ClashConfigsMergeServiceImpl(
    val clashConfigsMergeRepository: ClashConfigsMergeRepository,
    val clashConfigRepository: ClashConfigRepository
) : ClashConfigsMergeService {
    var template: String? = null

    @PostConstruct
    fun init() {
        val resource = ClassPathResource("config-template.json")
        resource.inputStream.use { inputStream ->
            template = StreamUtils.copyToString(inputStream, StandardCharsets.UTF_8)
        }
    }


    override fun create(name: String, configIds: MutableList<String>?): ClashConfigsMerge {
        val ccm = ClashConfigsMerge()
        ccm.name = name
        ccm.id = UUID.randomUUID().toString()
        if (configIds?.isNotEmpty() == true) {
            val css = clashConfigRepository.findAllById(configIds)
            ccm.configs = css
        }
        ccm.config = template
        ccm.token = UUID.randomUUID().toString()
        ccm.createdAt = Date()
        ccm.updatedAt = Date()
        return clashConfigsMergeRepository.save(ccm)
    }

    override fun save(clashConfigsMerge: ClashConfigsMerge): ClashConfigsMerge {
        val cids = clashConfigsMerge.configs?.map { c -> c.id }?.toList()
        if (cids?.isNotEmpty() == true) {
            clashConfigsMerge.configs = clashConfigRepository.findAllById(cids)
        }
        if (clashConfigsMerge.id?.isNotEmpty() == true) {
            val opt = clashConfigsMergeRepository.findById(clashConfigsMerge.id!!)
            opt.ifPresent {
                clashConfigsMerge.token = it.token
                clashConfigsMerge.createdAt = it.createdAt
                clashConfigsMerge.updatedAt = Date()
            }
        } else {
            clashConfigsMerge.id = UUID.randomUUID().toString()
            clashConfigsMerge.token = UUID.randomUUID().toString()
            clashConfigsMerge.updatedAt = Date()
            clashConfigsMerge.createdAt = Date()
        }
        return clashConfigsMergeRepository.save(clashConfigsMerge)
    }

    override fun detail(id: String): ClashConfigsMerge {
        return clashConfigsMergeRepository.findById(id).orElseThrow {
            return@orElseThrow ResponseStatusException(HttpStatus.NOT_FOUND)
        }
    }

    @Transactional
    override fun refreshToken(id: String): ClashConfigsMerge {
        var optCcm = clashConfigsMergeRepository
            .findById(id)
        if (optCcm.isEmpty) {
            throw ResponseStatusException(HttpStatus.NOT_FOUND)
        }
        val get = optCcm.get()
        get.token = UUID.randomUUID().toString()
        return clashConfigsMergeRepository.save(get)
    }

    override fun findByToken(token: String): ClashConfigsMerge? {
        return clashConfigsMergeRepository.findByToken(token)
    }

}