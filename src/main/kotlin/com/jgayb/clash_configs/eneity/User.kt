package com.jgayb.clash_configs.eneity

import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.Table

@Entity
@Table(name = "`user`")
data class User(
    @Id
    var id: String? = null,
    var username: String? = null,
    var password: String? = null,
)
