package com.jgayb.clash_configs

import com.jgayb.clash_configs.config.NativeHints
import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication
import org.springframework.context.annotation.ImportRuntimeHints
import org.springframework.scheduling.annotation.EnableScheduling

@SpringBootApplication
@EnableScheduling
@ImportRuntimeHints(NativeHints::class)
class ClashConfigsApplication

fun main(args: Array<String>) {
    runApplication<ClashConfigsApplication>(*args)
}
