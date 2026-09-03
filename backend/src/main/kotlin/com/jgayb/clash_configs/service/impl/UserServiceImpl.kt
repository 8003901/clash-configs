package com.jgayb.clash_configs.service.impl

import com.jgayb.clash_configs.eneity.User
import com.jgayb.clash_configs.repository.UserRepository
import com.jgayb.clash_configs.service.UserService
import jakarta.annotation.PostConstruct
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.core.env.Environment
import org.springframework.http.HttpStatus
import org.springframework.security.core.GrantedAuthority
import org.springframework.security.core.authority.SimpleGrantedAuthority
import org.springframework.security.core.userdetails.UserDetails
import org.springframework.security.core.userdetails.UsernameNotFoundException
import org.springframework.security.crypto.password.PasswordEncoder
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import org.springframework.web.server.ResponseStatusException
import java.util.*

@Service
class UserServiceImpl(
    val userRepository: UserRepository,
    val passwordEncoder: PasswordEncoder,
    var environment: Environment
) : UserService {

    private var log: Logger = LoggerFactory.getLogger(this.javaClass)

    @PostConstruct
    fun init() {
        var user = userRepository.findByUsername("admin")
        if (user == null) {
            user = userRepository.save(
                User(
                    id = UUID.randomUUID().toString(),
                    username = "admin",
                    password = passwordEncoder.encode("password")
                )
            )
        }
        if (environment.getProperty("pwdInit", "false") == "true") {
            user.password = passwordEncoder.encode("password")
            userRepository.save(user)
        }
    }

    override fun loadUserByUsername(username: String): UserDetails? {
        var user = userRepository.findByUsername(username)

        if (user == null) {
            log.warn("用户名：[{}]不存在", username)
            throw UsernameNotFoundException("用户名不存在")
        }

        return object : UserDetails {
            override fun getAuthorities(): Collection<GrantedAuthority?>? {
                return setOf(SimpleGrantedAuthority("ADMIN"))
            }

            override fun getPassword(): String? {
                return user.password
            }

            override fun getUsername(): String? {
                return user.username
            }

        }
    }

    @Transactional
    override fun changePwd(username: String, pwd: String, oldPwd: String) {
        val it = userRepository.findByUsername(username) ?: throw UsernameNotFoundException(username)
        if (!passwordEncoder.matches(oldPwd, it.password)) {
            throw ResponseStatusException(HttpStatus.BAD_REQUEST, "旧密码错误")
        }
        it.password = passwordEncoder.encode(pwd)
        userRepository.save(it)
    }
}