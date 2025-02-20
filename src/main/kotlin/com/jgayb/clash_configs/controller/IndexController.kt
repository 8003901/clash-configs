package com.jgayb.clash_configs.controller

import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RestController

@RestController
class IndexController {

    @GetMapping("/", "")
    fun home(): String {
        return "Hello Clash!"
    }

}