<script setup>
import { useRoute } from 'vue-router'

const route = useRoute()

function active(path) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}
</script>

<template>
  <el-container style="height: 100vh">
    <el-header class="hdr">
      <div class="brand">合同台账与付款节点提醒</div>
      <el-menu mode="horizontal" :ellipsis="false" router class="nav">
        <el-menu-item index="/" :class="{ 'is-active': active('/') && route.path === '/' }">工作台</el-menu-item>
        <el-menu-item index="/contracts" :class="{ 'is-active': active('/contracts') }">合同台账</el-menu-item>
        <el-menu-item index="/reports" :class="{ 'is-active': route.path.startsWith('/reports') }">
          合同报表
        </el-menu-item>
      </el-menu>
    </el-header>
    <el-main class="main">
      <router-view />
    </el-main>
  </el-container>
</template>

<style>
html,
body {
  margin: 0;
  font-family:
    ui-sans-serif,
    system-ui,
    -apple-system,
    'Segoe UI',
    Roboto,
    'PingFang SC',
    'Microsoft YaHei',
    sans-serif;
  background: #f4f6f9;
}

.hdr {
  display: flex;
  align-items: center;
  gap: 24px;
  border-bottom: 1px solid #ebeef5;
  background: #fff;
}
.brand {
  font-weight: 600;
  font-size: 17px;
  color: #303133;
  white-space: nowrap;
}
.nav {
  flex: 1;
  border-bottom: none !important;
}
.main {
  max-width: 1280px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}
</style>
