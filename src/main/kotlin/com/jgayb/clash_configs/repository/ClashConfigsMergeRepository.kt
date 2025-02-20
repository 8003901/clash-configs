package com.jgayb.clash_configs.repository

import com.jgayb.clash_configs.eneity.ClashConfigsMerge
import org.springframework.data.jpa.repository.JpaRepository

interface ClashConfigsMergeRepository : JpaRepository<ClashConfigsMerge, String> {

    fun findByToken(token: String): ClashConfigsMerge?

}