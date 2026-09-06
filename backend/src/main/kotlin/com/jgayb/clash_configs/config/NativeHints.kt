package com.jgayb.clash_configs.config

import org.springframework.aot.hint.RuntimeHints
import org.springframework.aot.hint.RuntimeHintsRegistrar
import org.springframework.core.io.ClassPathResource

/**
 * GraalVM native-image 运行时提示。
 *
 * 实体、DTO、控制器、服务等反射元数据已由 Spring Boot AOT 自动生成，无需在此注册；
 * 这里只需处理那些不在 static/ 下、运行时通过 [ClassPathResource] 手动加载的资源，
 * 否则 native-image 不会将其打包进可执行文件（启动时 FileNotFoundException）。
 */
class NativeHints : RuntimeHintsRegistrar {
    override fun registerHints(hints: RuntimeHints, classLoader: ClassLoader?) {
        hints.resources().registerResource(ClassPathResource("config-template.json"))
    }
}
