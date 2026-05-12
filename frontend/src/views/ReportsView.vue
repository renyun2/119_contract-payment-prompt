<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, shallowRef } from 'vue'
import * as echarts from 'echarts'
import api from '../api.js'
import { CONTRACT_TYPES } from '../constants.js'

const summary = ref(null)
const err = ref('')
const pieDiv = shallowRef(null)
const barDiv = shallowRef(null)
let pieChart
let barChart

function dispose() {
	pieChart?.dispose()
	barChart?.dispose()
	pieChart = barChart = null
}

async function load() {
	err.value = ''
	dispose()
	try {
		const { data } = await api.get('/reports/summary')
		summary.value = data
		await nextTick()
		renderCharts(data)
	} catch (e) {
		err.value = e?.response?.data?.error || e.message || '加载失败'
	}
}

function zhType(t) {
	return CONTRACT_TYPES.find((x) => x.value === t)?.label || t
}

function renderCharts(d) {
	if (pieDiv.value) {
		pieChart = echarts.init(pieDiv.value)
		pieChart.setOption({
			tooltip: { trigger: 'item', formatter: '{b}: ¥{c} ({d}%)' },
			legend: { bottom: 0 },
			series: [
				{
					type: 'pie',
					radius: ['36%', '62%'],
					data: (d.byType || []).map((x) => ({ name: zhType(x.type), value: Number(x.amount) })),
					label: { formatter: '{b}\n¥{c}' },
				},
			],
		})
	}
	if (barDiv.value) {
		barChart = echarts.init(barDiv.value)
		const y = d.year?.year || new Date().getFullYear()
		barChart.setOption({
			title: { text: y + ' 年收款 vs 付款（演示累计）', left: 'center', textStyle: { fontSize: 14 } },
			tooltip: { trigger: 'axis' },
			xAxis: { type: 'category', data: ['收款（销售）', '付款（采购/服务/工程等）'] },
			yAxis: { type: 'value', name: '元' },
			series: [
				{
					type: 'bar',
					barWidth: 48,
					data: [
						{ value: d.year?.income || 0, itemStyle: { color: '#67C23A' } },
						{ value: d.year?.expense || 0, itemStyle: { color: '#E6A23C' } },
					],
				},
			],
		})
	}
}

function onResize() {
	pieChart?.resize()
	barChart?.resize()
}

onMounted(() => {
	window.addEventListener('resize', onResize)
	load()
})

onBeforeUnmount(() => {
	window.removeEventListener('resize', onResize)
	dispose()
})
</script>

<template>
  <div>
    <el-page-header icon="" title="合同报表">
      <template #content><span class="ph-title">统计看板（截止 {{ summary?.asOf || '…' }}）</span></template>
      <template #extra>
        <el-button type="primary" @click="load">刷新</el-button>
      </template>
    </el-page-header>
    <el-alert v-if="err" type="error" show-icon style="margin: 12px 0">{{ err }}</el-alert>

    <el-row :gutter="16" style="margin-top: 12px">
      <el-col :xs="24" :md="12" :lg="6">
        <el-card shadow="hover">
          <div class="k">合同总数</div>
          <div class="v">{{ summary?.contracts?.total ?? '—' }}</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12" :lg="6">
        <el-card shadow="hover">
          <div class="k">履行中</div>
          <div class="v">{{ summary?.contracts?.active ?? '—' }}</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12" :lg="6">
        <el-card shadow="hover">
          <div class="k">已完成</div>
          <div class="v">{{ summary?.contracts?.completed ?? '—' }}</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12" :lg="6">
        <el-card shadow="hover">
          <div class="k">已终止</div>
          <div class="v">{{ summary?.contracts?.terminated ?? '—' }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :xs="24" :md="12">
        <el-card shadow="never">
          <template #header>本月应付侧（不含销售）</template>
          <p>计划<strong>本期未付</strong>：<span class="num">{{ summary?.payable?.dueUnpaidAmount?.toLocaleString?.() }}</span> 元</p>
          <p><strong>本期已付款</strong>流水：<span class="num">{{ summary?.payable?.paidAmount?.toLocaleString?.() }}</span> 元</p>
          <small class="hint">未付额为「计划中且未核销」的金额快照；已付额为「本月实际付款凭证」之和。</small>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card shadow="never">
          <template #header>本月应收侧（销售合同）</template>
          <p>计划<strong>本期未收回</strong>：<span class="num">{{ summary?.receivable?.dueUncollectedAmount?.toLocaleString?.() }}</span> 元</p>
          <p><strong>本期已回款</strong>：<span class="num">{{ summary?.receivable?.collectedAmount?.toLocaleString?.() }}</span> 元</p>
          <small class="hint">对应销售节点的计划收款日与逾期催收可在合同详情跟进。</small>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :xs="24" :md="14">
        <el-card shadow="never">
          <template #header>合同金额分布（按类型）</template>
          <div ref="pieDiv" style="height: 360px"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="10">
        <el-card shadow="never">
          <template #header>年度收支对比</template>
          <div ref="barDiv" style="height: 360px"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.ph-title {
  font-weight: 600;
  font-size: 17px;
}
.k {
  font-size: 13px;
  color: #909399;
}
.v {
  font-size: 26px;
  font-weight: 700;
  color: #303133;
  margin-top: 6px;
}
.num {
  font-weight: 600;
  color: #409eff;
}
.hint {
  color: #909399;
  line-height: 1.4;
}
</style>
