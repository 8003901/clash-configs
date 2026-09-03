package com.jgayb.clash_configs.service

import org.springframework.security.core.userdetails.UserDetailsService

interface UserService : UserDetailsService {

    fun changePwd(id: String, newPwd: String, oldPwd: String)

}