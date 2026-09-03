package com.jgayb.clash_configs.config

import jakarta.servlet.http.HttpServletRequest
import jakarta.servlet.http.HttpServletResponse
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.security.config.annotation.web.builders.HttpSecurity
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder
import org.springframework.security.crypto.password.PasswordEncoder
import org.springframework.security.web.SecurityFilterChain
import org.springframework.security.web.util.matcher.AntPathRequestMatcher

@Configuration
open class SecurityConfig {

    @Bean
    open fun securityFilterChain(http: HttpSecurity): SecurityFilterChain {
        http
            .authorizeHttpRequests { authorize ->
                authorize
                    .requestMatchers(AntPathRequestMatcher("/configs/**", "GET")).permitAll()
                    .requestMatchers("/", "/index.html", "/assets/**", "/favicon.svg", "/icons.svg").permitAll()
                    .anyRequest().authenticated()
            }
            .exceptionHandling { handling ->
                handling.authenticationEntryPoint { request, response, _ ->
                    if (isAjaxRequest(request)) {
                        response.status = HttpServletResponse.SC_UNAUTHORIZED
                        response.contentType = "application/json;charset=UTF-8"
                        response.writer.write("""{"error":"Unauthorized"}""")
                    } else {
                        response.sendRedirect("/login")
                    }
                }
            }
            .formLogin { form ->
                form
//                    .loginPage("/login.html")
//                    .loginProcessingUrl("/login")
                    .defaultSuccessUrl("/")
                    .permitAll()
            }
            .csrf { it.disable() }

        return http.build()
    }

    /**
     * 判断是否为前端 XHR/AJAX 请求。
     * 前端 axios 实例统一携带 X-Requested-With: XMLHttpRequest 与 Accept: application/json，
     * 对这类请求应返回 401（而非 302 重定向到 /login），由前端路由守卫拦截并跳转登录页。
     */
    private fun isAjaxRequest(request: HttpServletRequest): Boolean {
        val requestedWith = request.getHeader("X-Requested-With")
        if ("XMLHttpRequest".equals(requestedWith, ignoreCase = true)) {
            return true
        }
        val accept = request.getHeader("Accept")
        return accept != null && accept.contains("application/json")
    }

//    @Bean
//    fun userDetailsService(): UserDetailsService {
//        val userDetails = User.builder()
//            .username("admin")
//            .password(passwordEncoder().encode("admin"))
//            .roles("ADMIN")
//            .build()
//
//        return InMemoryUserDetailsManager(userDetails)
//    }

    @Bean
    open fun passwordEncoder(): PasswordEncoder {
        return BCryptPasswordEncoder()
    }
} 