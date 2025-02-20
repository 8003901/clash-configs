package com.jgayb.clash_configs.config

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
                    .requestMatchers("/").permitAll()
                    .anyRequest().authenticated()
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