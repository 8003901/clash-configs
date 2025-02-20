package com.jgayb.clash_configs.service

import com.jgayb.clash_configs.eneity.ClashConfigsMerge

interface ClashConfigsMergeService {

    fun create(name: String, configIds: MutableList<String>?): ClashConfigsMerge

    fun save(clashConfigsMerge: ClashConfigsMerge): ClashConfigsMerge

    fun refreshToken(id: String): ClashConfigsMerge

    fun findByToken(token: String): ClashConfigsMerge?

}