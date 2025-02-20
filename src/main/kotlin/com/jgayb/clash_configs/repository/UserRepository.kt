package com.jgayb.clash_configs.repository

import com.jgayb.clash_configs.eneity.User
import org.springframework.data.jpa.repository.JpaRepository

interface UserRepository : JpaRepository<User, String> {

    fun findByUsername(username: String): User?

}
