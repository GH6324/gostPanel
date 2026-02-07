<template>
  <div class="page-container">
    <div class="page-header">
      <h3>规则管理</h3>
    </div>
    <el-card shadow="hover">
      <!-- 搜索栏 -->
      <div class="search-bar">
        <div class="filters">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索名称或目标"
            :prefix-icon="Search"
            clearable
            style="width: 250px"
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          />
          <el-select v-model="searchNodeId" placeholder="选择节点" clearable style="width: 180px" @change="handleSearch">
            <el-option v-for="node in nodeList" :key="node.id" :label="node.name" :value="node.id" />
          </el-select>
          <el-select v-model="searchType" placeholder="规则类型" clearable style="width: 140px" @change="handleSearch">
            <el-option label="端口转发" value="forward" />
            <el-option label="隧道转发" value="tunnel" />
          </el-select>
          <el-select v-model="searchStatus" placeholder="状态" clearable style="width: 120px" @change="handleSearch">
            <el-option label="运行中" value="running" />
            <el-option label="已停止" value="stopped" />
          </el-select>
          <el-button :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="fetchData">刷新</el-button>
        </div>
        <div class="action-buttons">
          <el-button type="success" :icon="Download" @click="handleExport">导出</el-button>
          <el-button type="warning" :icon="Upload" @click="showImportDialog">导入</el-button>
          <el-button type="primary" :icon="Plus" @click="openDialog()">添加规则</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <el-table :data="ruleList" v-loading="loading" style="width: 100%" border>
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column prop="name" label="规则名" min-width="130" align="center" show-overflow-tooltip />
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.type === 'tunnel' ? 'warning' : 'primary'" size="small">
              {{ row.type === 'tunnel' ? '隧道转发' : '端口转发' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="入口" width="120" align="center">
          <template #default="{ row }">
            <template v-if="row.type === 'tunnel'">
              <el-tag size="small" type="warning">{{ row.tunnel?.entry_node?.name || '-' }}</el-tag>
            </template>
            <template v-else>
              <el-tag size="small" type="primary">{{ row.node?.name || '-' }}</el-tag>
            </template>
          </template>
        </el-table-column>
        <el-table-column label="隧道" width="120" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.tunnel" size="small" type="warning">{{ row.tunnel?.name || '-' }}</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="protocol" label="协议" width="80" align="center">
          <template #default="{ row }">
            <el-tag size="small">{{ (row.protocol || 'tcp').toUpperCase() }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="listen_port" label="监听端口" width="100" align="center" />
        <el-table-column label="目标地址" min-width="150" align="center" show-overflow-tooltip>
          <template #default="{ row }">
              <span v-if="row.targets && row.targets.length > 0">{{ row.targets[0] }}<span v-if="row.targets.length > 1"> (+{{ row.targets.length - 1 }})</span></span>
              <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="总流量" width="120" align="center">
          <template #default="{ row }">
            {{ formatBytes(row.total_bytes || 0) }}
          </template>
        </el-table-column>
        <el-table-column label="上传流量" width="120" align="center">
          <template #default="{ row }">
             <span style="color: #67c23a">{{ formatBytes(row.input_bytes || 0) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="下载流量" width="120" align="center">
          <template #default="{ row }">
             <span style="color: #409eff">{{ formatBytes(row.output_bytes || 0) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" align="center" fixed="right">
          <template #default="{ row }">
            <el-button 
              v-if="row.status !== 'running'" 
              type="success" link size="small" 
              @click="handleStart(row)"
            >启动</el-button>
            <el-button 
              v-else 
              type="warning" link size="small" 
              @click="handleStop(row)"
            >停止</el-button>
            <el-button type="primary" link size="small" @click="openDialog(row)">编辑</el-button>
            <el-button type="info" link size="small" @click="handleCopy(row)">复制</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchData"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑规则' : '添加规则'"
      width="650px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入规则名称" :prefix-icon="EditPen" />
        </el-form-item>
        <el-form-item label="规则类型" prop="type">
          <el-select v-model="form.type" :disabled="isEdit" @change="handleTypeChange" style="width: 100%">
            <el-option label="端口转发" value="forward" />
            <el-option label="隧道转发" value="tunnel" />
          </el-select>
          <div class="form-hint">端口转发：选择节点直接转发 | 隧道转发：选择隧道通过链路转发</div>
        </el-form-item>
        <!-- 端口转发：选择入口节点 -->
        <el-form-item v-if="form.type === 'forward'" label="入口节点" prop="node_id">
          <el-select v-model="form.node_id" placeholder="请选择入口节点" style="width: 100%" :disabled="isEdit">
            <el-option v-for="node in nodeList" :key="node.id" :label="node.name" :value="node.id" />
          </el-select>
          <div class="form-hint">直接在该节点上创建转发服务</div>
        </el-form-item>
        <!-- 隧道转发：选择隧道 -->
        <el-form-item v-if="form.type === 'tunnel'" label="选择隧道" prop="tunnel_id">
          <el-select v-model="form.tunnel_id" placeholder="请选择隧道" style="width: 100%" :disabled="isEdit">
            <el-option 
              v-for="tunnel in tunnelList" 
              :key="tunnel.id" 
              :label="`${tunnel.name} (${tunnel.entry_node?.name || '-'} → ${tunnel.exit_node?.name || '-'})`" 
              :value="tunnel.id" 
            />
          </el-select>
          <div class="form-hint">在隧道的入口节点上创建转发服务，流量通过隧道链路转发</div>
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="协议" prop="protocol">
              <el-select v-model="form.protocol" placeholder="选择协议" style="width: 100%">
                <el-option label="TCP" value="tcp" />
                <el-option label="UDP" value="udp" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="监听端口" prop="listen_port">
              <el-input-number v-model="form.listen_port" :min="1" :max="65535" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="负载均衡" prop="strategy">
               <el-select v-model="form.strategy" placeholder="默认为轮询">
                  <el-option label="轮询" value="round"/>
                  <el-option label="随机" value="rand"/>
                  <el-option label="先进先出" value="fifo"/>
                  <el-option label="哈希" value="hash"/>
               </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="目标列表" style="margin-bottom: 0;">
           <el-table :data="form.targetList" border style="width: 100%" size="small" :show-header="true">
              <el-table-column label="目标地址 (IP:Port)" min-width="250">
                  <template #default="{ row }">
                      <el-input v-model="row.address" placeholder="例如: 192.168.1.100:8080" />
                  </template>
              </el-table-column>
              <el-table-column label="操作" width="60" align="center">
                  <template #default="{ $index }">
                      <el-button type="danger" link :icon="UseRemove" @click="removeTarget($index)" />
                  </template>
              </el-table-column>
           </el-table>
           <div style="margin-top: 10px; text-align: center; width: 100%;">
               <el-button type="primary" link :icon="Plus" @click="addTarget" style="width: 100%; border: 1px dashed #dcdfe6;">添加目标地址</el-button>
           </div>
        </el-form-item>
        
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="备注信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 导出弹框 -->
    <el-dialog v-model="exportDialogVisible" title="导出规则" width="600px">
      <el-alert type="info" :closable="false" style="margin-bottom: 16px;">
        复制以下 JSON 数据，可用于导入到其他面板
      </el-alert>
      <el-input
        v-model="exportJsonData"
        type="textarea"
        :rows="15"
        readonly
      />
      <template #footer>
        <el-button @click="exportDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="copyExportData">复制数据</el-button>
      </template>
    </el-dialog>

    <!-- 导入弹框 -->
    <el-dialog v-model="importDialogVisible" title="导入规则" width="600px">
      <el-alert type="warning" :closable="false" style="margin-bottom: 16px;">
        粘贴 JSON 数据，将批量创建规则（需要手动修改端口避免冲突）
      </el-alert>
      <el-input
        v-model="importJsonData"
        type="textarea"
        :rows="15"
        placeholder="请粘贴 JSON 数据..."
      />
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="importLoading" @click="handleImport">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search, EditPen, Remove as UseRemove, Download, Upload } from '@element-plus/icons-vue'
import { getRuleList, createRule, updateRule, deleteRule, startRule, stopRule } from '@/api/rule'
import { getNodeList } from '@/api/node'
import { getTunnelList } from '@/api/tunnel'

// 节点列表
const nodeList = ref([])
// 隧道列表
const tunnelList = ref([])

// 列表数据
const ruleList = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 搜索
const searchKeyword = ref('')
const searchNodeId = ref('')
const searchType = ref('')
const searchStatus = ref('')

// 对话框
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const submitLoading = ref(false)
const formRef = ref(null)

// 导入导出弹框
const exportDialogVisible = ref(false)
const exportJsonData = ref('')
const importDialogVisible = ref(false)
const importJsonData = ref('')
const importLoading = ref(false)

const form = reactive({
  type: 'forward',
  node_id: '',
  tunnel_id: null,
  name: '',
  protocol: 'tcp',
  listen_port: 0,
  targetList: [{ address: '' }],
  strategy: 'round',
  remark: ''
})

// 动态验证规则
const validateEntry = (rule, value, callback) => {
  if (form.type === 'forward' && !form.node_id) {
    callback(new Error('请选择入口节点'))
  } else if (form.type === 'tunnel' && !form.tunnel_id) {
    callback(new Error('请选择隧道'))
  } else {
    callback()
  }
}

const formRules = {
  type: [{ required: true, message: '请选择规则类型', trigger: 'change' }],
  node_id: [{ validator: validateEntry, trigger: 'change' }],
  tunnel_id: [{ validator: validateEntry, trigger: 'change' }],
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  protocol: [{ required: true, message: '请选择协议', trigger: 'change' }],
  listen_port: [{ required: true, message: '请输入监听端口', trigger: 'blur' }]
}

// 状态处理
const getStatusType = (status) => {
  const map = { running: 'success', stopped: 'info', error: 'danger' }
  return map[status] || 'info'
}

const getStatusText = (status) => {
  const map = { running: '运行中', stopped: '已停止', error: '错误' }
  return map[status] || status
}

// 格式化字节数
const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}

// 获取节点列表
const fetchNodes = async () => {
  try {
    const res = await getNodeList({ pageSize: 100 })
    nodeList.value = res.data.list || []
  } catch (error) {
    console.error('获取节点列表失败:', error)
  }
}

// 获取隧道列表
const fetchTunnels = async () => {
  try {
    const res = await getTunnelList({ pageSize: 100 })
    tunnelList.value = res.data.list || []
  } catch (error) {
    console.error('获取隧道列表失败:', error)
  }
}

// 获取数据
const fetchData = async (isSilent = false) => {
  if (!isSilent) loading.value = true
  try {
    const res = await getRuleList({
      page: page.value,
      pageSize: pageSize.value,
      node_id: searchNodeId.value,
      type: searchType.value,
      status: searchStatus.value,
      keyword: searchKeyword.value
    })
    ruleList.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    console.error('获取规则列表失败:', error)
  } finally {
    if (!isSilent) loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  page.value = 1
  fetchData()
}

// 打开对话框
const openDialog = (row = null) => {
  isEdit.value = !!row
  editId.value = row?.id || null
  
  if (row) {
    // 解析 targets
    let tList = []
    if (row.targets && row.targets.length > 0) {
        tList = row.targets.map(t => ({ address: t }))
    }

    Object.assign(form, {
      type: row.type || 'forward',
      node_id: row.node_id,
      tunnel_id: row.tunnel_id || null,
      name: row.name,
      protocol: row.protocol,
      listen_port: row.listen_port,
      targetList: tList.length > 0 ? tList : [{ address: '' }],
      strategy: row.strategy || 'round',
      remark: row.remark || ''
    })
  } else {
    Object.assign(form, {
      type: 'forward',
      node_id: '',
      tunnel_id: null,
      name: '',
      protocol: 'tcp',
      listen_port: 8000,
      targetList: [{ address: '' }],
      strategy: 'round',
      remark: ''
    })
  }
  
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitLoading.value = true
    try {
      // 准备提交数据
      const targets = form.targetList.map(item => item.address).filter(t => t.trim() !== '')
      
      const submitData = {
        type: form.type,
        node_id: form.type === 'forward' ? form.node_id : null,
        tunnel_id: form.type === 'tunnel' ? form.tunnel_id : null,
        name: form.name,
        protocol: form.protocol,
        listen_port: form.listen_port,
        targets: targets,
        strategy: form.strategy,
        remark: form.remark
      }
      
      if (isEdit.value) {
        await updateRule(editId.value, submitData)
        ElMessage.success('更新成功')
      } else {
        await createRule(submitData)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      fetchData()
    } catch (error) {
      console.error('操作失败:', error)
    } finally {
      submitLoading.value = false
    }
  })
}

// 删除规则
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要删除规则 "${row.name}" 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteRule(row.id)
    ElMessage.success('删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
    }
  }
}

// 启动规则
const handleStart = async (row) => {
  try {
    await startRule(row.id)
    ElMessage.success('启动成功')
    fetchData()
  } catch (error) {
    console.error('启动失败:', error)
  }
}

// 停止规则
const handleStop = async (row) => {
  try {
    await stopRule(row.id)
    ElMessage.success('停止成功')
    fetchData()
  } catch (error) {
    console.error('停止失败:', error)
  }
}

// 复制规则
const handleCopy = (row) => {
  isEdit.value = false
  editId.value = null
  
  // 解析 targets
  let tList = []
  if (row.targets && row.targets.length > 0) {
    tList = row.targets.map(t => ({ address: t }))
  }

  Object.assign(form, {
    type: row.type || 'forward',
    node_id: row.node_id,
    tunnel_id: row.tunnel_id || null,
    name: row.name + '-copy',
    protocol: row.protocol,
    listen_port: row.listen_port + 1, // 端口自动+1
    targetList: tList.length > 0 ? tList : [{ address: '' }],
    strategy: row.strategy || 'round',
    remark: row.remark || ''
  })
  
  dialogVisible.value = true
}

// 导出规则
const handleExport = async () => {
  try {
    // 获取全部规则
    const res = await getRuleList({ pageSize: 1000 })
    const allRules = res.data.list || []
    
    if (allRules.length === 0) {
      ElMessage.warning('没有可导出的规则')
      return
    }
    
    // 整理导出数据，只保留核心字段
    const exportData = allRules.map(rule => ({
      name: rule.name,
      type: rule.type,
      node_id: rule.node_id,
      tunnel_id: rule.tunnel_id,
      protocol: rule.protocol,
      listen_port: rule.listen_port,
      targets: rule.targets,
      strategy: rule.strategy,
      remark: rule.remark || ''
    }))
    
    exportJsonData.value = JSON.stringify(exportData, null, 2)
    exportDialogVisible.value = true
  } catch (error) {
    console.error('获取规则失败:', error)
    ElMessage.error('获取规则失败')
  }
}

// 复制导出数据
const copyExportData = async () => {
  try {
    await navigator.clipboard.writeText(exportJsonData.value)
    ElMessage.success('已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

// 显示导入弹框
const showImportDialog = () => {
  importJsonData.value = ''
  importDialogVisible.value = true
}

// 导入规则
const handleImport = async () => {
  if (!importJsonData.value.trim()) {
    ElMessage.warning('请粘贴 JSON 数据')
    return
  }
  
  let rules = []
  try {
    rules = JSON.parse(importJsonData.value)
    if (!Array.isArray(rules)) {
      rules = [rules]
    }
  } catch (error) {
    ElMessage.error('JSON 格式错误')
    return
  }
  
  if (rules.length === 0) {
    ElMessage.warning('没有可导入的规则')
    return
  }
  
  importLoading.value = true
  let successCount = 0
  let failCount = 0
  
  for (const rule of rules) {
    try {
      await createRule({
        name: rule.name || '导入规则',
        type: rule.type || 'forward',
        node_id: rule.node_id,
        tunnel_id: rule.tunnel_id,
        protocol: rule.protocol || 'tcp',
        listen_port: rule.listen_port || 0,
        targets: rule.targets || [],
        strategy: rule.strategy || 'round',
        remark: rule.remark || ''
      })
      successCount++
    } catch (error) {
      failCount++
      console.error('导入规则失败:', error)
    }
  }
  
  importLoading.value = false
  importDialogVisible.value = false
  
  if (successCount > 0) {
    ElMessage.success(`成功导入 ${successCount} 条规则` + (failCount > 0 ? `，失败 ${failCount} 条` : ''))
    fetchData()
  } else {
    ElMessage.error('导入失败')
  }
}

// 定时刷新
let refreshTimer = null

onMounted(() => {
  fetchNodes()
  fetchTunnels()
  fetchData()
  
  // 每 5 秒刷新一次 (静默刷新)
  refreshTimer = setInterval(() => {
    fetchData(true)
  }, 5000)
})

onBeforeUnmount(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})

// 添加目标
const addTarget = () => {
    form.targetList.push({ address: '' })
}

// 移除目标
const removeTarget = (index) => {
    form.targetList.splice(index, 1)
}

// 切换规则类型时清空另一侧的选择
const handleTypeChange = (val) => {
  if (val === 'forward') {
    form.tunnel_id = null
  } else {
    form.node_id = ''
  }
}
</script>

<style scoped>
.page-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header h3 {
  margin: 0 0 16px 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.search-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.filters {
  display: flex;
  gap: 12px;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.text-muted {
  color: #909399;
  font-size: 12px;
}

.form-hint {
  color: #909399;
  font-size: 12px;
  margin-top: 4px;
}

:deep(.el-table .el-table__cell) {
  padding: 12px 0;
}
</style>
