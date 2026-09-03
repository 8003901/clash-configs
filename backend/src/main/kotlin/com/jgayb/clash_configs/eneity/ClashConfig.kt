package com.jgayb.clash_configs.eneity

import jakarta.persistence.*
import org.springframework.data.annotation.CreatedDate
import org.springframework.data.annotation.LastModifiedDate
import java.util.*

@Entity
@Table(name = "clash_configs")
data class ClashConfig(
    @Id
    var id: String? = null,
    var url: String? = null,
    var name: String? = null,
    var enabled: Boolean = true,
    @Enumerated(EnumType.STRING)
    var updateSchedule: UpdateSchedule = UpdateSchedule.DAY,
    @Column(length = 500000)
    var content: String? = null,
    @CreatedDate
    var createdAt: Date? = null,
    @LastModifiedDate
    var updatedAt: Date? = null,
    var subscriptionUserinfo: String? = null,
)

enum class UpdateSchedule {
    DAY,
    WEEK
}