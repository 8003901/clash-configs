<script setup lang="ts">
import { Toaster } from '@/components/ui/sonner'
</script>

<template>
  <RouterView />
  <Toaster position="top-right" :offset="16" />
</template>
