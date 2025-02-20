package com.jgayb.clash_configs.controller

import com.jgayb.clash_configs.service.UserService
import org.springframework.security.core.context.SecurityContextHolder
import org.springframework.security.core.userdetails.UserDetails
import org.springframework.web.bind.annotation.PutMapping
import org.springframework.web.bind.annotation.RestController

@RestController
class UserController(val userService: UserService) {

    @PutMapping("/password")
    fun changePassword(oldPassword: String, newPassword: String) {
        var ud = SecurityContextHolder.getContext().authentication.principal as UserDetails
        userService.changePwd(ud.username, oldPassword, newPassword)
    }

}