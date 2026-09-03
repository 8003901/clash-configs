package com.jgayb.clash_configs

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication
import org.springframework.scheduling.annotation.EnableScheduling

@SpringBootApplication
@EnableScheduling
class ClashConfigsApplication

fun main(args: Array<String>) {
    runApplication<ClashConfigsApplication>(*args)
}
