<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api.js'
import { labelStatus, labelType } from '../constants.js'
import { ElMessage, ElMessageBox } from 'element-plus'

const route = useRoute()
const contractId = computed(() => route.params.id)

const contract = ref(null)
const nodes = ref([])
const payments = ref([])
const followups = ref([])
const tab = ref('nodes')
const loading = ref(true)

const nodeDialog = ref(false)
const editingNode = ref(null)
const nodeForm = reactive({
  nodeName: '',
  triggerCondition: '',
  amount: 0,
  plannedDate: '',
  isTriggered: false,
})

const payDialog = ref(false)
const payForm = reactive({
  nodeId: null,
  payDate: '',
  amount: 0,
  bankRef: '',
  payAccount: '',
})

const fuDialog = ref(false)
const fuForm = reactive({
  nodeId: null,
  follower: '',
  followDate: '',
  content: '',
  promisedPayDate: '',
})

async function loadAll() {
  loading.value = true
  try {
    const id = contractId.value
    const [c, n, p, f] = await Promise.all([
      api.get('/contracts/' + id),
      api.get('/contracts/' + id + '/nodes'),
      api.get('/payments', { params: { contractId: id } }),
      api.get('/contracts/' + id + '/followups').catch(() => ({ data: [] })),
    ])
    contract.value = c.data
    nodes.value = n.data
    payments.value = p.data
    followups.value = Array.isArray(f.data) ? f.data : []
    if (contract.value?.type === 'sales') tab.value = 'nodes'
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadAll)
watch(contractId, loadAll)

function openNodeDialog(row) {
  editingNode.value = row || null
  if (row) {
    nodeForm.nodeName = row.nodeName
    nodeForm.triggerCondition = row.triggerCondition
    nodeForm.amount = row.amount
    nodeForm.plannedDate = row.plannedDate
    nodeForm.isTriggered = row.isTriggered
  } else {
    nodeForm.nodeName = ''
    nodeForm.triggerCondition = ''
    nodeForm.amount = 0
    nodeForm.plannedDate = ''
    nodeForm.isTriggered = false
  }
  nodeDialog.value = true
}

async function saveNode() {
  const id = contractId.value
  const body = {
    nodeName: nodeForm.nodeName,
    triggerCondition: nodeForm.triggerCondition,
    amount: nodeForm.amount,
    plannedDate: nodeForm.plannedDate,
    isTriggered: nodeForm.isTriggered,
  }
  try {
    if (editingNode.value) {
      await api.put(`/contracts/${id}/nodes/${editingNode.value.id}`, body)
    } else {
      await api.post(`/contracts/${id}/nodes`, body)
    }
    nodeDialog.value = false
    await loadAll()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '保存失败')
  }
}

async function delNode(row) {
  try {
    await ElMessageBox.confirm('删除节点「' + row.nodeName + '」？', '确认', { type: 'warning' })
  } catch {
    return
  }
  await api.delete(`/contracts/${contractId.value}/nodes/${row.id}`)
  await loadAll()
}

const unpaidNodes = computed(() => nodes.value.filter((n) => !n.isPaid))

function openPay() {
  payForm.nodeId = unpaidNodes.value[0]?.id || null
  payForm.payDate = new Date().toISOString().slice(0, 10)
  payForm.amount = 0
  payForm.bankRef = ''
  payForm.payAccount = ''
  payDialog.value = true
}

async function savePay() {
  const id = Number(contractId.value)
  try {
    await api.post('/payments', {
      contractId: id,
      nodeId: payForm.nodeId,
      payDate: payForm.payDate,
      amount: payForm.amount || 0,
      bankRef: payForm.bankRef,
      payAccount: payForm.payAccount,
    })
    payDialog.value = false
    ElMessage.success('已登记实际付款，节点已置为已付')
    await loadAll()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '登记失败')
  }
}

function openFollow() {
  fuForm.nodeId = null
  fuForm.follower = ''
  fuForm.followDate = new Date().toISOString().slice(0, 10)
  fuForm.content = ''
  fuForm.promisedPayDate = ''
  fuDialog.value = true
}

async function saveFollow() {
  const id = contractId.value
  try {
    await api.post(`/contracts/${id}/followups`, {
      nodeId: fuForm.nodeId || undefined,
      follower: fuForm.follower,
      followDate: fuForm.followDate,
      content: fuForm.content,
      promisedPayDate: fuForm.promisedPayDate || undefined,
    })
    fuDialog.value = false
    ElMessage.success('催收记录已保存')
    await loadAll()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '保存失败')
  }
}

function rowNodeClass({ row }) {
  if (row.alertLevel === 'red') return 'cp-row-red'
  if (row.alertLevel === 'orange') return 'cp-row-orange'
  return ''
}
</script>

<template>
  <div v-loading="loading">
    <el-page-header @back="$router.push('/contracts')" title="返回列表">
      <template #content>
        <span class="ph-title">{{ contract?.title || '合同详情' }}</span>
      </template>
    </el-page-header>

    <el-descriptions v-if="contract" :column="3" border style="margin-top: 12px" size="small">
      <el-descriptions-item label="编号">{{ contract.contractNo }}</el-descriptions-item>
      <el-descriptions-item label="类型">{{ labelType(contract.type) }}</el-descriptions-item>
      <el-descriptions-item label="状态">{{ labelStatus(contract.status) }}</el-descriptions-item>
      <el-descriptions-item label="对方单位" :span="2">{{ contract.counterparty }}</el-descriptions-item>
      <el-descriptions-item label="签订日">{{ String(contract.signedDate).slice(0, 10) }}</el-descriptions-item>
      <el-descriptions-item label="总金额">
        {{ Number(contract.totalAmount).toLocaleString(undefined, { minimumFractionDigits: 2 }) }}
      </el-descriptions-item>
      <el-descriptions-item label="期限起">{{ contract.periodStart ? String(contract.periodStart).slice(0, 10) : '—' }}</el-descriptions-item>
      <el-descriptions-item label="期限止">{{ contract.periodEnd ? String(contract.periodEnd).slice(0, 10) : '—' }}</el-descriptions-item>
      <el-descriptions-item label="摘要" :span="3">{{ contract.summary || '—' }}</el-descriptions-item>
    </el-descriptions>

    <el-tabs v-model="tab" style="margin-top: 16px">
      <el-tab-pane label="付款/收款计划" name="nodes">
        <div class="toolbar">
          <el-button type="primary" @click="openNodeDialog(null)">新增节点</el-button>
          <el-button :disabled="unpaidNodes.length === 0" @click="openPay">登记实际付款</el-button>
        </div>
        <el-table :data="nodes" stripe :row-class-name="rowNodeClass" style="margin-top: 12px">
          <el-table-column prop="nodeName" label="节点" width="100" />
          <el-table-column prop="triggerCondition" label="触发条件" min-width="200" show-overflow-tooltip />
          <el-table-column prop="amount" label="金额" width="118">
            <template #default="{ row }">{{ Number(row.amount).toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</template>
          </el-table-column>
          <el-table-column prop="plannedDate" label="计划日" width="110" />
          <el-table-column label="已触发" width="80">
            <template #default="{ row }">{{ row.isTriggered ? '是' : '否' }}</template>
          </el-table-column>
          <el-table-column label="已付/已收" width="88">
            <template #default="{ row }">{{ row.isPaid ? '是' : '否' }}</template>
          </el-table-column>
          <el-table-column label="预警" width="120">
            <template #default="{ row }">
              <el-tag v-if="row.alertLevel === 'red'" type="danger">红色</el-tag>
              <el-tag v-else-if="row.alertLevel === 'orange'" type="warning">橙色</el-tag>
              <span v-else>—</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openNodeDialog(row)">编辑</el-button>
              <el-button link type="danger" @click="delNode(row)">删</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="实际流水" name="pay">
        <el-table :data="payments" stripe>
          <el-table-column prop="payDate" label="付款日" width="110">
            <template #default="{ row }">{{ String(row.payDate).slice(0, 10) }}</template>
          </el-table-column>
          <el-table-column prop="amount" label="金额" width="120">
            <template #default="{ row }">{{ Number(row.amount).toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</template>
          </el-table-column>
          <el-table-column prop="bankRef" label="银行流水号" min-width="160" />
          <el-table-column prop="payAccount" label="付款账户" min-width="160" />
        </el-table>
      </el-tab-pane>

      <el-tab-pane v-if="contract?.type === 'sales'" label="催收跟进" name="fu">
        <el-button type="primary" @click="openFollow">新增跟进</el-button>
        <el-timeline style="margin-top: 16px">
          <el-timeline-item v-for="f in followups" :key="f.id" :timestamp="String(f.followDate).slice(0, 10)" placement="top">
            <p>
              <b>{{ f.follower }}</b>
              <span v-if="f.promisedPayDate"> · 承诺付款日 {{ String(f.promisedPayDate).slice(0, 10) }}</span>
            </p>
            <p class="fu-content">{{ f.content }}</p>
          </el-timeline-item>
        </el-timeline>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="nodeDialog" :title="editingNode ? '编辑节点' : '新增节点'" width="520px">
      <el-form label-width="108px">
        <el-form-item label="节点名称" required>
          <el-select v-model="nodeForm.nodeName" allow-create filterable default-first-option placeholder="或输入自定义" style="width: 100%">
            <el-option label="预付款" value="预付款" />
            <el-option label="进度款" value="进度款" />
            <el-option label="尾款" value="尾款" />
            <el-option label="质保金" value="质保金" />
          </el-select>
        </el-form-item>
        <el-form-item label="触发条件">
          <el-input v-model="nodeForm.triggerCondition" type="textarea" rows="2" />
        </el-form-item>
        <el-form-item label="应付/应收额">
          <el-input-number v-model="nodeForm.amount" :min="0" :precision="2" :step="1000" style="width: 100%" />
        </el-form-item>
        <el-form-item label="计划日期" required>
          <el-date-picker v-model="nodeForm.plannedDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="已触发">
          <el-switch v-model="nodeForm.isTriggered" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="nodeDialog = false">取消</el-button>
        <el-button type="primary" @click="saveNode">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="payDialog" title="登记实际付款" width="480px">
      <el-form label-width="110px">
        <el-form-item label="关联节点" required>
          <el-select v-model="payForm.nodeId" style="width: 100%" placeholder="请选择未付节点">
            <el-option v-for="u in unpaidNodes" :key="u.id" :label="u.nodeName + '（' + u.plannedDate + '）'" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="实际付款日" required>
          <el-date-picker v-model="payForm.payDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="金额">
          <el-input-number v-model="payForm.amount" :min="0" :precision="2" style="width: 100%" />
          <small class="hint">留空或 0 则默认节点全额</small>
        </el-form-item>
        <el-form-item label="银行流水号"><el-input v-model="payForm.bankRef" /></el-form-item>
        <el-form-item label="付款账户"><el-input v-model="payForm.payAccount" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="payDialog = false">取消</el-button>
        <el-button type="primary" @click="savePay">确认</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="fuDialog" title="催收跟进记录" width="540px">
      <el-alert type="info" show-icon style="margin-bottom: 12px" :closable="false">
        销售合同逾期未收时使用；记录跟进人、沟通内容及承诺回款日。
      </el-alert>
      <el-form label-width="120px">
        <el-form-item label="关联节点（可选）">
          <el-select v-model="fuForm.nodeId" clearable placeholder="可不选">
            <el-option v-for="u in unpaidNodes" :key="u.id" :label="u.nodeName" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="跟进人" required><el-input v-model="fuForm.follower" /></el-form-item>
        <el-form-item label="跟进日期" required>
          <el-date-picker v-model="fuForm.followDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="沟通内容" required><el-input v-model="fuForm.content" type="textarea" rows="4" /></el-form-item>
        <el-form-item label="承诺付款日">
          <el-date-picker v-model="fuForm.promisedPayDate" type="date" value-format="YYYY-MM-DD" clearable style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="fuDialog = false">取消</el-button>
        <el-button type="primary" @click="saveFollow">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.ph-title {
  font-weight: 600;
}
.toolbar {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.fu-content {
  color: #606266;
  margin: 4px 0 0;
  white-space: pre-wrap;
}
.hint {
  display: block;
  color: #909399;
  margin-top: 4px;
}
:deep(.cp-row-red td) {
  background-color: rgba(245, 108, 108, 0.1) !important;
}
:deep(.cp-row-orange td) {
  background-color: rgba(230, 162, 60, 0.12) !important;
}
</style>
