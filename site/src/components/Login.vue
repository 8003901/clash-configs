<template>
  <div class="login-container">
    <v-sheet class="mx-auto" width="340" style="padding-left: 20px; padding-right: 20px;">
      <div style="padding-top: 20px; padding-bottom: 40px; display: flex; justify-content: space-evenly;">
        <img :src="logo" alt="log" width="200px">
      </div>
      <v-form @submit.prevent style="padding-bottom: 20px;">
        <v-text-field @keydown.enter="handleEnter" v-model="state.username"
          :error-messages="v$.username.$errors.map(e => e.$message)" label="Username" required
          @blur="v$.username.$touch" @input="v$.username.$touch" />
        <v-text-field @keydown.enter="handleEnter" v-model="state.password"
          :error-messages="v$.password.$errors.map(e => e.$message)" label="Password" required
          @blur="v$.password.$touch" @input="v$.password.$touch" :type="'password'" />
        <div class="login-button-container">
          <v-btn class="mt-4" @click="loginFun" color="blue" width="100" :disabled="isDisabled">登录</v-btn>
          <!-- <v-btn class="mt-4" @click="clear" width="100" color="orange">
          clear
        </v-btn> -->
        </div>
      </v-form>
    </v-sheet>
  </div>
</template>
<script setup>
import { ref, reactive } from 'vue';
import { useVuelidate } from '@vuelidate/core'
import { helpers, required } from '@vuelidate/validators'
import logo from '../assets/keketata.png'
import service from '@/plugins/request'
import { validate } from '@babel/types'
import router from '@/router'
import { jwtDecode } from 'jwt-decode'
import { useToast } from 'vue-toastification'

const toast = useToast()

const isDisabled = ref(false)

const jwt = localStorage.getItem('token')
if (jwt != null && jwt != '') {
  const exp = jwtDecode(jwt).exp
  const now = (new Date().getTime() / 1000)
  if (exp >= now) {
    loginSuccess()
  }
}

function loginSuccess() {
  toast.success('登录成功！', {
    timeout: 2000,
  })
  router.replace('/')
}


const initialState = {
  username: '',
  password: '',
  show1: false,
}
const state = reactive({
  ...initialState,
})
const rules = {
  username: { required: helpers.withMessage('用户名不能为空', required) },
  password: { required: helpers.withMessage('密码不能为空', required) },
}
const v$ = useVuelidate(rules, state)

function handleEnter() {
  loginFun()
}

async function loginFun() {
  isDisabled.value = true
  const vr = await v$.value.$validate()
  if (vr) {
    const resp = await service.post("/api/login", state, {
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
    })
    const jwt = resp.jwt
    localStorage.setItem('token', jwt)
    loginSuccess()
    // console.log(resp)
  }
  isDisabled.value = false
}

function clear() {
  v$.value.$reset()

  for (const [key, value] of Object.entries(initialState)) {
    state[key] = value
  }
}
</script>
<style>
.login-button-container {
  display: flex;
  justify-content: space-evenly;
  /* 使按钮之间的间隔均匀分布 */
}

.login-container {
  margin-top: 200px;
}
</style>
