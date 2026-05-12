<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api.js'
import { CONTRACT_STATUS, CONTRACT_TYPES, labelStatus, labelType } from '../constants.js'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()
const loading = ref(true)
const table = ref([])
const q = reactive({ keyword: '', type: '', status: '' })
const dialogVisible = ref(false)
const saving = ref(false)
const form = reactive({
  id: null,
  contractNo: '',
  title: '',
  signedDate: '',
  type: 'purchase',
  counterparty: '',
  totalAmount: 0,
  periodStart: '',
  periodEnd: '',
  summary: '',
  status: 'active',
})

async function load() {
  loading.value = true
  try {
    const params = {}
    if (q.keyword) params.q = q.keyword
    if (q.type) params.type = q.type
    if (q.status) params.status = q.status
    const { data } = await api.get('/contracts', { params })
    table.value = data
  } finally {
    loading.value = false
  }
}

onMounted(load)

function openCreate() {
  form.id = null
  form.contractNo = ''
  form.title = ''
  form.signedDate = ''
  form.type = 'purchase'
  form.counterparty = ''
  form.totalAmount = 0
  form.periodStart = ''
  form.periodEnd = ''
  form.summary = ''
  form.status = 'active'
  dialogVisible.value = true
}

async function openEdit(row) {
  try {
    const { data } = await api.get('/contracts/' + row.id)
    form.id = data.id
    form.contractNo = data.contractNo
    form.title = data.title
    form.signedDate = data.signedDate?.slice?.(0, 10) ?? data.signedDate
    form.type = data.type
    form.counterparty = data.counterparty
    form.totalAmount = data.totalAmount
    form.periodStart = data.periodStart ? data.periodStart.slice(0, 10) : ''
    form.periodEnd = data.periodEnd ? data.periodEnd.slice(0, 10) : ''
    form.summary = data.summary || ''
    form.status = data.status
    dialogVisible.value = true
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '加载合同失败')
  }
}

async function save() {
  saving.value = true
  try {
    const body = {
      contractNo: form.contractNo,
      title: form.title,
      signedDate: form.signedDate,
      type: form.type,
      counterparty: form.counterparty,
      totalAmount: form.totalAmount,
      summary: form.summary,
      status: form.status,
    }
    if (form.periodStart) body.periodStart = form.periodStart
    if (form.periodEnd) body.periodEnd = form.periodEnd

    if (form.id) await api.put('/contracts/' + form.id, body)
    else await api.post('/contracts', body)
    dialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

async function removeRow(row) {
  try {
    await ElMessageBox.confirm('确定删除合同「' + row.title + '」？将级联删除付款节点与流水。', '删除', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await api.delete('/contracts/' + row.id)
    ElMessage.success('已删除')
    await load()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '删除失败')
  }
}

</script>

<template>
  <div>
    <el-page-header icon="" title="合同台账">
      <template #content>
        <span class="ph-title">录入与维护</span>
      </template>
      <template #extra>
        <el-button type="primary" @click="openCreate">新建合同</el-button>
      </template>
    </el-page-header>

    <div class="toolbar">
      <el-input v-model="q.keyword" clearable placeholder="名称/编号/对方单位" style="width: 240px" />
      <el-select v-model="q.type" clearable placeholder="类型" style="width: 120px">
        <el-option v-for="o in CONTRACT_TYPES" :key="o.value" :label="o.label" :value="o.value" />
      </el-select>
      <el-select v-model="q.status" clearable placeholder="状态" style="width: 120px">
        <el-option v-for="o in CONTRACT_STATUS" :key="o.value" :label="o.label" :value="o.value" />
      </el-select>
      <el-button @click="load">查询</el-button>
    </div>

    <el-table v-loading="loading" :data="table" stripe style="width: 100%; margin-top: 12px">
      <el-table-column prop="contractNo" label="合同编号" width="130" />
      <el-table-column prop="title" label="合同名称" min-width="200" />
      <el-table-column label="类型" width="80">
        <template #default="{ row }">{{ labelType(row.type) }}</template>
      </el-table-column>
      <el-table-column prop="counterparty" label="对方单位" width="140" />
      <el-table-column prop="signedDate" label="签订日" width="110">
        <template #default="{ row }">{{ String(row.signedDate).slice(0, 10) }}</template>
      </el-table-column>
      <el-table-column prop="totalAmount" label="总金额" width="120">
        <template #default="{ row }">{{ Number(row.totalAmount).toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="88">
        <template #default="{ row }">{{ labelStatus(row.status) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="router.push('/contracts/' + row.id)">详情</el-button>
          <el-button link @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="removeRow(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑合同' : '新建合同'" width="640px">
      <el-form label-width="112px">
        <el-form-item label="合同编号" required>
          <el-input v-model="form.contractNo" :disabled="!!form.id" autocomplete="off" />
        </el-form-item>
        <el-form-item label="合同名称" required>
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="签订日期" required>
          <el-date-picker v-model="form.signedDate" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" />
        </el-form-item>
        <el-form-item label="合同类型" required>
          <el-select v-model="form.type" style="width: 100%">
            <el-option v-for="o in CONTRACT_TYPES" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="对方单位" required>
          <el-input v-model="form.counterparty" />
        </el-form-item>
        <el-form-item label="合同总金额">
          <el-input-number v-model="form.totalAmount" :precision="2" :step="1000" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="合同期限起">
          <el-date-picker v-model="form.periodStart" type="date" value-format="YYYY-MM-DD" clearable placeholder="可选" />
        </el-form-item>
        <el-form-item label="合同期限止">
          <el-date-picker v-model="form.periodEnd" type="date" value-format="YYYY-MM-DD" clearable placeholder="可选" />
        </el-form-item>
        <el-form-item label="主要内容">
          <el-input v-model="form.summary" type="textarea" rows="3" />
        </el-form-item>
        <el-form-item label="合同状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option v-for="o in CONTRACT_STATUS" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.ph-title {
  font-weight: 600;
  font-size: 17px;
}
.toolbar {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
  margin-top: 14px;
}
</style>
