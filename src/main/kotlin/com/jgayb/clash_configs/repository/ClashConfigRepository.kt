package com.jgayb.clash_configs.repository

import com.jgayb.clash_configs.eneity.ClashConfig
import org.springframework.data.jpa.repository.JpaRepository

interface ClashConfigRepository : JpaRepository<ClashConfig, String> {
}