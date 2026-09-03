package com.jgayb.clash_configs.controller

import org.springframework.stereotype.Controller
import org.springframework.web.bind.annotation.GetMapping

@Controller
class IndexController {

    @GetMapping("/", "")
    fun index(): String = "forward:/index.html"

}