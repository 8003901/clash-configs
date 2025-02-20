package com.jgayb.clash_configs.repository

import com.jgayb.clash_configs.eneity.ClashConfig
import com.jgayb.clash_configs.eneity.UpdateSchedule
import jakarta.persistence.QueryHint
import org.hibernate.jpa.HibernateHints.HINT_FETCH_SIZE
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.data.jpa.repository.Query
import org.springframework.data.jpa.repository.QueryHints
import org.springframework.data.repository.query.Param
import java.util.stream.Stream


interface ClashConfigRepository : JpaRepository<ClashConfig, String> {

    @Query("select t from ClashConfig t where t.updateSchedule = :fre")
    @QueryHints(value = [QueryHint(name = HINT_FETCH_SIZE, value = "" + Int.MIN_VALUE)])
    fun streamAll(@Param("fre") fre: UpdateSchedule): Stream<ClashConfig>?

}