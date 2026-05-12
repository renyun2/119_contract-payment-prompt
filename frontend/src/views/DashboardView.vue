<script setup>
import { onMounted, ref } from 'vue'
import api from '../api.js'
import { labelType } from '../constants.js'

const loading = ref(true)
const data = ref(null)
const err = ref('')

async function load() {
  loading.value = true
  err.value = ''
  try {
    const { data: body } = await api.get('/dashboard/upcoming')
    data.value = body
  } catch (e) {
    err.value = e?.response?.data?.error || e.message || '加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(load)

function rowClass({ row }) {
  if (row.alertLevel === 'red') return 'cp-row-red'
  if (row.alertLevel === 'orange') return 'cp-row-orange'
  return ''
}

function tagType(level) {
  if (level === 'red') return 'danger'
  if (level === 'orange') return 'warning'
  return 'success'
}

function tagText(level, contractType) {
  const base = contractType === 'sales' ? '应收' : '应付'
  if (level === 'red') return base + '·红色预警'
  if (level === 'orange') return base + '·橙色预警'
  return '正常'
}
</script>

<template>
  <div>
    <el-page-header icon="" title="工作台">
      <template #content>
        <span class="ph-title">本月到期应付/应收节点汇总</span>
      </template>
      <template #extra>
        <el-button type="primary" @click="load">刷新</el-button>
      </template>
    </el-page-header>
    <p v-if="data" class="sub">
      统计区间：<b>{{ data.rangeFrom }}</b> ～ <b>{{ data.rangeTo }}</b>
      ，今日 <b>{{ data.today }}</b>。未到期的<strong>应付</strong>节点在到期前 15 天标橙、7 天内标红；<strong>销售合同应收</strong>逾期未收标红，并支持催收跟进。
    </p>
    <el-alert v-if="err" :title="err" type="error" show-icon style="margin-bottom: 12px" />
    <el-skeleton :loading="loading" animated :rows="6">
      <el-table :data="data?.items || []" stripe :row-class-name="rowClass" style="width: 100%">
        <el-table-column prop="contractTitle" label="合同名称" min-width="200" />
        <el-table-column prop="counterparty" label="对方单位" width="160" />
        <el-table-column label="类型" width="88">
          <template #default="{ row }">{{ labelType(row.contractType) }}</template>
        </el-table-column>
        <el-table-column prop="nodeName" label="节点" width="100" />
        <el-table-column prop="amount" label="应付/应收额" width="120">
          <template #default="{ row }">{{ Number(row.amount).toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</template>
        </el-table-column>
        <el-table-column prop="plannedDate" label="计划日" width="110" />
        <el-table-column label="触发" width="72">
          <template #default="{ row }">{{ row.isTriggered ? '已触发' : '未触发' }}</template>
        </el-table-column>
        <el-table-column label="预警" width="118">
          <template #default="{ row }">
            <el-tag v-if="row.alertLevel !== 'none'" :type="tagType(row.alertLevel)" disable-transitions>
              {{ tagText(row.alertLevel, row.contractType) }}
            </el-tag>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="108" fixed="right">
          <template #default="{ row }">
            <router-link :to="'/contracts/' + row.contractId">
              <el-button link type="primary">进入合同</el-button>
            </router-link>
          </template>
        </el-table-column>
      </el-table>
    </el-skeleton>
  </div>
</template>

<style scoped>
.ph-title {
  font-weight: 600;
  font-size: 17px;
}
.sub {
  color: #606266;
  font-size: 13px;
  line-height: 1.55;
  margin: 12px 0 16px;
}
:deep(.cp-row-red td) {
  background-color: rgba(245, 108, 108, 0.12) !important;
}
:deep(.cp-row-orange td) {
  background-color: rgba(230, 162, 60, 0.15) !important;
}
</style>
