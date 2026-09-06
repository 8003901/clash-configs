package com.jgayb.clash_configs.controller

import com.jgayb.clash_configs.service.UserService
import org.springframework.security.core.context.SecurityContextHolder
import org.springframework.security.core.userdetails.UserDetails
import org.springframework.web.bind.annotation.PutMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RestController

@RestController
class UserController(val userService: UserService) {

    @PutMapping("/users/me/password")
    fun changePassword(@RequestBody passwordUpdate: PasswordUpdate) {
        val authentication = requireNotNull(SecurityContextHolder.getContext().authentication) {
            "Request has no authentication"
        }
        val ud = authentication.principal as UserDetails
        val username = requireNotNull(ud.username) { "Authenticated user has no username" }
        userService.changePwd(username, passwordUpdate.newPassword, passwordUpdate.oldPassword)
    }

}

data class PasswordUpdate(val oldPassword: String, val newPassword: String)