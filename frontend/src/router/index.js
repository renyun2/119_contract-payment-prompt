import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '../views/DashboardView.vue'
import ContractsView from '../views/ContractsView.vue'
import ContractDetailView from '../views/ContractDetailView.vue'
import ReportsView from '../views/ReportsView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'dashboard', component: DashboardView },
    { path: '/contracts', name: 'contracts', component: ContractsView },
    { path: '/contracts/:id', name: 'contractDetail', component: ContractDetailView, props: true },
    { path: '/reports', name: 'reports', component: ReportsView },
  ],
})

export default router
