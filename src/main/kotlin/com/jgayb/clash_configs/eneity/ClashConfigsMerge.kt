package com.jgayb.clash_configs.eneity

import jakarta.persistence.*
import org.hibernate.annotations.DynamicUpdate
import org.springframework.data.annotation.CreatedDate
import org.springframework.data.annotation.LastModifiedDate
import java.util.*

@Entity
@Table(name = "clash_configs_merge")
@DynamicUpdate
data class ClashConfigsMerge(
    @Id
    var id: String? = null,
    var name: String? = null,
    var token: String? = null,
    @Column(length = 50000)
    var config: String? = null,
    @CreatedDate
    var createdAt: Date? = null,
    @LastModifiedDate
    var updatedAt: Date? = null,

    @ManyToMany(fetch = FetchType.LAZY)
    @JoinTable(
        name = "clash_configs_merge_config",
        joinColumns = [JoinColumn(
            name = "clash_configs_merge_id",
            foreignKey = ForeignKey(name = "none", value = ConstraintMode.NO_CONSTRAINT)
        )],
        inverseJoinColumns = [JoinColumn(
            name = "config_id",
            foreignKey = ForeignKey(name = "none", value = ConstraintMode.NO_CONSTRAINT)
        )]
    )
    var configs: MutableList<ClashConfig>? = null
)
