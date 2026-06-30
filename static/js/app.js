// API 配置
const API_BASE_URL = '/api/v1';

// 当前用户信息
let currentUser = null;
let authToken = localStorage.getItem('token');

// 清空表单
function clearForm() {
    document.getElementById('long-url-input').value = '';
    document.getElementById('link-name').value = '';
    document.getElementById('link-remark').value = '';
    document.getElementById('expire-days').value = '7';
    
    // 隐藏清除图标
    updateClearIcon('long-url-input');
    updateClearIcon('link-name');
    updateClearIcon('link-remark');
    
    // 隐藏结果区域
    const resultDiv = document.getElementById('result');
    if (resultDiv) {
        resultDiv.classList.add('hidden');
    }
}

// 清空单个输入框
function clearInput(inputId) {
    const input = document.getElementById(inputId);
    if (input) {
        input.value = '';
        updateClearIcon(inputId);
        input.focus();
    }
}

// 更新清除图标显示状态
function updateClearIcon(inputId) {
    const input = document.getElementById(inputId);
    const clearBtn = document.getElementById('clear-' + inputId);
    if (input && clearBtn) {
        clearBtn.style.display = input.value.trim() ? 'flex' : 'none';
    }
}

// 更新有效期选项（根据登录状态）
function updateExpireOptions() {
    const select = document.getElementById('expire-days');
    const isLoggedIn = currentUser !== null;
    
    // 清空现有选项
    select.innerHTML = '';
    
    // 游客：只显示一周内选项
    // 登录用户/管理员：显示全部选项（含永久）
    const options = isLoggedIn 
        ? [
            { value: 1, text: '1天' },
            { value: 3, text: '3天' },
            { value: 7, text: '1周', selected: true },
            { value: 30, text: '1个月' },
            { value: 365, text: '1年' },
            { value: 0, text: '永久（不过期）' }
          ]
        : [
            { value: 1, text: '1天' },
            { value: 3, text: '3天' },
            { value: 7, text: '1周', selected: true }
          ];
    
    options.forEach(opt => {
        const option = document.createElement('option');
        option.value = opt.value;
        option.textContent = opt.text;
        if (opt.selected) {
            option.selected = true;
        }
        select.appendChild(option);
    });
}

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    updateStatsCards();
    checkLoginStatus();
    
    // 为输入框添加事件监听
    const inputs = ['long-url-input', 'link-name', 'link-remark'];
    inputs.forEach(id => {
        const input = document.getElementById(id);
        if (input) {
            input.addEventListener('input', () => updateClearIcon(id));
            input.addEventListener('keyup', () => updateClearIcon(id));
        }
    });
});

// ==================== 基础 API ====================

// 创建短链
async function createShortURL(longURL, expireDays, customDomain, name, remark) {
    const response = await fetch(`${API_BASE_URL}/short-url`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            long_url: longURL,
            expire_days: expireDays,
            custom_domain: customDomain,
            name: name,
            remark: remark
        })
    });

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '创建短链失败');
    }

    return response.json();
}

// 获取访问统计
async function getStats(shortCode) {
    const response = await fetch(`${API_BASE_URL}/short-url/${shortCode}/stats`);
    
    if (!response.ok) {
        throw new Error('获取统计信息失败');
    }

    return response.json();
}

// 获取短链列表
async function getShortURLList(page = 1, limit = 10) {
    const response = await fetch(`${API_BASE_URL}/short-url/list?page=${page}&limit=${limit}`);
    
    if (!response.ok) {
        throw new Error('获取短链列表失败');
    }

    return response.json();
}

// ==================== 认证 API ====================

// 用户登录
async function login(username, password) {
    const response = await fetch(`${API_BASE_URL}/auth/login`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            username: username,
            password: password
        })
    });

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '登录失败');
    }

    return response.json();
}

// 获取用户信息
async function getProfile() {
    const response = await fetch(`${API_BASE_URL}/user/profile`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${authToken}`
        }
    });

    if (!response.ok) {
        throw new Error('获取用户信息失败');
    }

    return response.json();
}

// 修改密码
async function changePassword(oldPassword, newPassword) {
    const response = await fetch(`${API_BASE_URL}/user/password`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${authToken}`
        },
        body: JSON.stringify({
            old_password: oldPassword,
            new_password: newPassword
        })
    });

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '修改密码失败');
    }

    return response.json();
}

// ==================== 管理员 API ====================

// 获取用户短链列表（普通用户）
async function getUserShortURLList(page = 1, pageSize = 10) {
    const response = await fetch(`${API_BASE_URL}/user/short-url/list?page=${page}&page_size=${pageSize}`, {
        headers: {
            'Authorization': `Bearer ${authToken}`
        }
    });

    if (!response.ok) {
        throw new Error('获取短链列表失败');
    }

    return response.json();
}

// 获取管理员短链列表
async function getAdminShortURLList(page = 1, pageSize = 10) {
    const response = await fetch(`${API_BASE_URL}/admin/short-url/list?page=${page}&page_size=${pageSize}`, {
        headers: {
            'Authorization': `Bearer ${authToken}`
        }
    });

    if (!response.ok) {
        throw new Error('获取短链列表失败');
    }

    return response.json();
}

// 删除短链 API
async function deleteShortURLAPI(code) {
    const response = await fetch(`${API_BASE_URL}/admin/short-url/${code}`, {
        method: 'DELETE',
        headers: {
            'Authorization': `Bearer ${authToken}`
        }
    });

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '删除失败');
    }

    return response.json();
}

// 更新短链状态
async function updateShortURLStatus(code, status) {
    const response = await fetch(`${API_BASE_URL}/admin/short-url/${code}/status`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${authToken}`
        },
        body: JSON.stringify({
            status: status
        })
    });

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '更新状态失败');
    }

    return response.json();
}

// 获取用户列表
async function getUserList(page = 1, pageSize = 10) {
    const response = await fetch(`${API_BASE_URL}/admin/users?page=${page}&page_size=${pageSize}`, {
        headers: {
            'Authorization': `Bearer ${authToken}`
        }
    });

    if (!response.ok) {
        throw new Error('获取用户列表失败');
    }

    return response.json();
}

// 删除用户
async function deleteUser(id) {
    const response = await fetch(`${API_BASE_URL}/admin/users/${id}`, {
        method: 'DELETE',
        headers: {
            'Authorization': `Bearer ${authToken}`
        }
    });

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '删除失败');
    }

    return response.json();
}

// 更新用户
async function updateUser(id, data) {
    const response = await fetch(`${API_BASE_URL}/admin/users/${id}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${authToken}`
        },
        body: JSON.stringify(data)
    });

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '更新失败');
    }

    return response.json();
}

// ==================== UI 功能 ====================

// 显示结果
function showResult(shortCode) {
    const resultDiv = document.getElementById('result');
    const resultUrlInput = document.getElementById('short-url-result');
    
    // 构建短链URL，使用当前页面的协议和域名
    const protocol = window.location.protocol;
    const domain = window.location.host;
    const shortURL = `${protocol}//${domain}/${shortCode}`;
    
    resultUrlInput.value = shortURL;
    resultDiv.classList.remove('hidden');
    
    // 自动更新统计信息
    updateStats(shortCode);
}

// 更新统计信息
async function updateStats(shortCode) {
    try {
        const stats = await getStats(shortCode);
        if (stats.code === 0) {
            document.getElementById('total-visits').textContent = stats.data.total_visits || 0;
            document.getElementById('today-visits').textContent = stats.data.today_visits || 0;
        }
    } catch (error) {
        console.error('更新统计信息失败:', error);
    }
}

// 更新统计卡片
async function updateStatsCards() {
    try {
        const list = await getShortURLList(1, 1000);
        if (list.code === 0) {
            document.getElementById('stats-total-count').textContent = list.data.total || 0;
            
            // 计算总访问次数
            let totalVisits = 0;
            list.data.items.forEach(item => {
                totalVisits += item.total_visits || 0;
            });
            document.getElementById('stats-total-visits').textContent = totalVisits;
        }
    } catch (error) {
        console.error('更新统计卡片失败:', error);
    }
}

// 复制到剪贴板
async function copyToClipboard(text) {
    try {
        await navigator.clipboard.writeText(text);
        alert('已复制到剪贴板！');
    } catch (error) {
        // 降级方案
        const textarea = document.createElement('textarea');
        textarea.value = text;
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand('copy');
        document.body.removeChild(textarea);
        alert('已复制到剪贴板！');
    }
}

// 检查登录状态
async function checkLoginStatus() {
    if (authToken) {
        try {
            const result = await getProfile();
            if (result.code === 0) {
                currentUser = result.data;
                updateNavbar();
                updateExpireOptions();
            }
        } catch (error) {
            console.error('检查登录状态失败:', error);
            logout();
        }
    } else {
        // 游客状态也要渲染有效期选项
        updateExpireOptions();
    }
}

// 更新导航栏
function updateNavbar() {
    document.getElementById('nav-login').classList.add('hidden');
    document.getElementById('nav-profile').classList.remove('hidden');
    document.getElementById('logout-btn').classList.remove('hidden');
    
    if (currentUser && currentUser.role === 1) {
        document.getElementById('nav-admin').classList.remove('hidden');
    }
}

// 退出登录
function logout() {
    localStorage.removeItem('token');
    authToken = null;
    currentUser = null;
    
    document.getElementById('nav-login').classList.remove('hidden');
    document.getElementById('nav-profile').classList.add('hidden');
    document.getElementById('nav-admin').classList.add('hidden');
    document.getElementById('logout-btn').classList.add('hidden');
    
    showPage('home');
    alert('已退出登录');
}

// 显示页面
function showPage(pageName) {
    // 隐藏所有页面
    document.querySelectorAll('.page').forEach(page => {
        page.classList.remove('active');
    });
    
    // 隐藏所有导航链接的active状态
    document.querySelectorAll('.nav-link').forEach(link => {
        link.classList.remove('active');
    });
    
    // 显示对应页面
    document.getElementById(`page-${pageName}`).classList.add('active');
    
    // 激活对应导航链接
    const navLink = document.querySelector(`[onclick="showPage('${pageName}')"]`);
    if (navLink) {
        navLink.classList.add('active');
    }
    
    // 如果是用户中心，加载用户信息
    if (pageName === 'profile') {
        loadProfile();
    }
    
    // 如果是管理后台，加载数据
    if (pageName === 'admin') {
        loadShortURLs();
    }
}

// 用户中心标签页切换
function showProfileTab(tabName) {
    document.querySelectorAll('.profile-tab').forEach(tab => {
        tab.classList.add('hidden');
    });
    document.querySelectorAll('.profile-tabs .tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    const targetTab = document.getElementById(`tab-${tabName}`);
    targetTab.classList.remove('hidden');
    
    const buttons = document.querySelectorAll('.profile-tabs .tab-btn');
    buttons.forEach(btn => {
        if (btn.getAttribute('onclick').includes(tabName)) {
            btn.classList.add('active');
        }
    });
    
    // 如果切换到我的短链标签页，加载短链列表
    if (tabName === 'mylinks') {
        loadMyShortURLs();
    }
}

// 显示管理后台标签页
function showAdminTab(tabName) {
    document.querySelectorAll('.admin-tab').forEach(tab => {
        tab.classList.remove('active');
        tab.classList.add('hidden');
    });
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    const targetTab = document.getElementById(`tab-${tabName}`);
    targetTab.classList.remove('hidden');
    targetTab.classList.add('active');
    document.querySelector(`[onclick="showAdminTab('${tabName}')"]`).classList.add('active');
    
    if (tabName === 'users') {
        loadUsers();
    } else if (tabName === 'shorturls') {
        loadShortURLs();
    }
}

// 加载用户信息
async function loadProfile() {
    try {
        const result = await getProfile();
        if (result.code === 0) {
            currentUser = result.data;
            document.getElementById('profile-username').textContent = currentUser.username || '-';
            document.getElementById('profile-email').textContent = currentUser.email || '-';
            document.getElementById('profile-role').textContent = currentUser.role === 1 ? '管理员' : '普通用户';
        }
    } catch (error) {
        console.error('加载用户信息失败:', error);
        alert('加载用户信息失败');
    }
}

// 加载短链列表
async function loadShortURLs() {
    try {
        const result = await getAdminShortURLList(1, 100);
        if (result.code === 0) {
            renderShortURLTable(result.data.items);
        }
    } catch (error) {
        console.error('加载短链列表失败:', error);
        alert('加载短链列表失败');
    }
}

// 加载我的短链列表
async function loadMyShortURLs() {
    try {
        const result = await getUserShortURLList(1, 100);
        if (result.code === 0) {
            renderMyShortURLTable(result.data.items);
        } else {
            console.error('加载我的短链列表失败:', result.message);
            alert('加载失败: ' + result.message);
        }
    } catch (error) {
        console.error('加载我的短链列表失败:', error);
        alert('加载短链列表失败');
    }
}

// 渲染我的短链表格
function renderMyShortURLTable(items) {
    const tbody = document.getElementById('mylinks-table-body');
    tbody.innerHTML = '';
    
    if (!items || items.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;">暂无数据</td></tr>';
        return;
    }
    
    items.forEach(item => {
        const shortURL = `http://${window.location.host}/${item.short_code}`;
        const remainingDays = calculateRemainingDays(item.expire_at);
        const row = document.createElement('tr');
        row.innerHTML = `
            <td><a href="${shortURL}" target="_blank">${item.short_code || '-'}</a></td>
            <td title="${item.name || ''}">${item.name || '-'}</td>
            <td title="${item.long_url}">${truncateText(item.long_url, 40)}</td>
            <td>${item.total_visits || 0}</td>
            <td>${remainingDays}</td>
            <td><span class="${item.status === 1 ? 'status-active' : 'status-disabled'}">${item.status === 1 ? '启用' : '禁用'}</span></td>
        `;
        tbody.appendChild(row);
    });
}

// 加载用户列表
async function loadUsers() {
    try {
        const result = await getUserList(1, 100);
        if (result.code === 0) {
            renderUserTable(result.data.items);
        }
    } catch (error) {
        console.error('加载用户列表失败:', error);
        alert('加载用户列表失败');
    }
}

// 渲染短链表格
function renderShortURLTable(items) {
    const tbody = document.getElementById('shorturls-table-body');
    tbody.innerHTML = '';
    
    if (!items || items.length === 0) {
        tbody.innerHTML = '<tr><td colspan="9" style="text-align:center;">暂无数据</td></tr>';
        return;
    }
    
    items.forEach(item => {
        const shortURL = `http://${window.location.host}/${item.short_code}`;
        const remainingDays = calculateRemainingDays(item.expire_at);
        const row = document.createElement('tr');
        row.innerHTML = `
            <td><a href="${shortURL}" target="_blank">${item.short_code || '-'}</a></td>
            <td title="${item.name || ''}">${item.name || '-'}</td>
            <td title="${item.remark || ''}">${truncateText(item.remark, 30) || '-'}</td>
            <td title="${item.long_url}">${truncateText(item.long_url, 40)}</td>
            <td>${item.total_visits || 0}</td>
            <td>${remainingDays}</td>
            <td>${item.created_by || 'guest'}</td>
            <td><span class="${item.status === 1 ? 'status-active' : 'status-disabled'}">${item.status === 1 ? '启用' : '禁用'}</span></td>
            <td>
                <button class="btn-small btn-copy" onclick="copyShortURL('${shortURL}')" title="复制短链">复制</button>
                <button class="btn-small btn-toggle" onclick="toggleShortURLStatus('${item.short_code}', ${item.status})">
                    ${item.status === 1 ? '禁用' : '启用'}
                </button>
                <button class="btn-small btn-delete" onclick="confirmDeleteShortURL('${item.short_code}')">删除</button>
            </td>
        `;
        tbody.appendChild(row);
    });
}

// 计算剩余有效天数
function calculateRemainingDays(expireAt) {
    if (!expireAt) {
        return '永久';
    }
    
    try {
        const expireDate = new Date(expireAt);
        const now = new Date();
        const diffTime = expireDate - now;
        
        if (diffTime <= 0) {
            return '<span style="color: red;">已过期</span>';
        }
        
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
        return diffDays;
    } catch (error) {
        console.error('计算剩余天数失败:', error);
        return '未知';
    }
}

// 渲染用户表格
function renderUserTable(items) {
    const tbody = document.getElementById('users-table-body');
    tbody.innerHTML = '';
    
    if (!items || items.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;">暂无数据</td></tr>';
        return;
    }
    
    items.forEach(item => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${item.id || '-'}</td>
            <td>${item.username || '-'}</td>
            <td>${item.email || '-'}</td>
            <td>${item.role === 1 ? '管理员' : '普通用户'}</td>
            <td><span class="${item.status === 1 ? 'status-active' : 'status-disabled'}">${item.status === 1 ? '启用' : '禁用'}</span></td>
            <td>
                <button class="btn-small btn-edit" onclick="editUser(${item.id})">编辑</button>
                <button class="btn-small btn-toggle" onclick="toggleUserStatus(${item.id}, ${item.status})">
                    ${item.status === 1 ? '禁用' : '启用'}
                </button>
                <button class="btn-small btn-delete" onclick="confirmDeleteUser(${item.id})">删除</button>
            </td>
        `;
        tbody.appendChild(row);
    });
}

// 搜索短链
function searchShortURLs() {
    const keyword = document.getElementById('search-shortcode').value.toLowerCase();
    const rows = document.querySelectorAll('#shorturls-table-body tr');
    
    rows.forEach(row => {
        const shortCode = row.querySelector('td:first-child').textContent.toLowerCase();
        const longURL = row.querySelector('td:nth-child(2)').textContent.toLowerCase();
        
        if (shortCode.includes(keyword) || longURL.includes(keyword)) {
            row.style.display = '';
        } else {
            row.style.display = 'none';
        }
    });
}

// 搜索用户
function searchUsers() {
    const keyword = document.getElementById('search-username').value.toLowerCase();
    const rows = document.querySelectorAll('#users-table-body tr');
    
    rows.forEach(row => {
        const username = row.querySelector('td:nth-child(2)').textContent.toLowerCase();
        const email = row.querySelector('td:nth-child(3)').textContent.toLowerCase();
        
        if (username.includes(keyword) || email.includes(keyword)) {
            row.style.display = '';
        } else {
            row.style.display = 'none';
        }
    });
}

// 切换短链状态
async function toggleShortURLStatus(code, currentStatus) {
    try {
        const newStatus = currentStatus === 1 ? 0 : 1;
        const result = await updateShortURLStatus(code, newStatus);
        
        if (result.code === 0) {
            loadShortURLs();
            alert('状态已更新');
        } else {
            alert(result.message || '更新失败');
        }
    } catch (error) {
        console.error('更新状态失败:', error);
        alert('更新状态失败');
    }
}

// 切换用户状态
async function toggleUserStatus(id, currentStatus) {
    try {
        const newStatus = currentStatus === 1 ? 0 : 1;
        const result = await updateUser(id, { status: newStatus });
        
        if (result.code === 0) {
            loadUsers();
            alert('状态已更新');
        } else {
            alert(result.message || '更新失败');
        }
    } catch (error) {
        console.error('更新状态失败:', error);
        alert('更新状态失败');
    }
}

// 复制短链
function copyShortURL(url) {
    try {
        navigator.clipboard.writeText(url).then(() => {
            alert('短链已复制到剪贴板！');
        }).catch(() => {
            // 降级方案
            const textarea = document.createElement('textarea');
            textarea.value = url;
            document.body.appendChild(textarea);
            textarea.select();
            document.execCommand('copy');
            document.body.removeChild(textarea);
            alert('短链已复制到剪贴板！');
        });
    } catch (error) {
        console.error('复制失败:', error);
        alert('复制失败');
    }
}

// 确认删除短链
function confirmDeleteShortURL(code) {
    if (confirm(`确定要删除短链 ${code} 吗？`)) {
        deleteShortURLConfirm(code);
    }
}

// 删除短链
async function deleteShortURLConfirm(code) {
    try {
        const result = await deleteShortURLAPI(code);
        
        if (result.code === 0) {
            loadShortURLs();
            alert('删除成功');
        } else {
            alert(result.message || '删除失败');
        }
    } catch (error) {
        console.error('删除失败:', error);
        alert('删除失败：' + error.message);
    }
}

// 确认删除用户
// 编辑用户
async function editUser(id) {
    try {
        const result = await getUser(id);
        if (result.code === 0) {
            showEditUserModal(result.data);
        } else {
            alert(result.message || '获取用户信息失败');
        }
    } catch (error) {
        console.error('获取用户信息失败:', error);
        alert('获取用户信息失败');
    }
}

// 获取单个用户
async function getUser(id) {
    const response = await fetch(`${API_BASE_URL}/admin/users/${id}`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${authToken}`
        }
    });
    
    if (!response.ok) {
        throw new Error('获取用户信息失败');
    }
    
    return response.json();
}

function confirmDeleteUser(id) {
    if (confirm(`确定要删除用户 ID: ${id} 吗？`)) {
        deleteUserConfirm(id);
    }
}

// 删除用户
async function deleteUserConfirm(id) {
    try {
        const result = await deleteUser(id);
        
        if (result.code === 0) {
            loadUsers();
            alert('删除成功');
        } else {
            alert(result.message || '删除失败');
        }
    } catch (error) {
        console.error('删除失败:', error);
        alert('删除失败');
    }
}

// 显示新建用户模态框
function showCreateUserModal() {
    document.getElementById('create-user-modal').classList.remove('hidden');
}

// 关闭新建用户模态框
function closeCreateUserModal() {
    document.getElementById('create-user-modal').classList.add('hidden');
    document.getElementById('create-user-form').reset();
}

// 显示编辑用户模态框
function showEditUserModal(user) {
    document.getElementById('edit-user-id').value = user.id;
    document.getElementById('edit-username').value = user.username;
    document.getElementById('edit-password').value = '';
    document.getElementById('edit-email').value = user.email || '';
    document.getElementById('edit-role').value = user.role;
    document.getElementById('edit-status').value = user.status;
    document.getElementById('edit-user-modal').classList.remove('hidden');
}

// 关闭编辑用户模态框
function closeEditUserModal() {
    document.getElementById('edit-user-modal').classList.add('hidden');
    document.getElementById('edit-user-form').reset();
}

// 创建用户
async function createUser(username, password, email, role, status) {
    const response = await fetch(`${API_BASE_URL}/admin/users`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${authToken}`
        },
        body: JSON.stringify({
            username: username,
            password: password,
            email: email,
            role: role,
            status: status
        })
    });

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || '创建失败');
    }

    return response.json();
}

// 表单提交处理
document.getElementById('create-user-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const username = document.getElementById('new-username').value.trim();
    const password = document.getElementById('create-password').value;
    const email = document.getElementById('new-email').value.trim();
    const role = parseInt(document.getElementById('new-role').value);
    const status = parseInt(document.getElementById('new-status').value);
    
    try {
        const result = await createUser(username, password, email, role, status);
        
        if (result.code === 0) {
            closeCreateUserModal();
            loadUsers();
            alert('用户创建成功');
        } else {
            alert(result.message || '创建失败');
        }
    } catch (error) {
        console.error('创建用户失败:', error);
        alert('创建用户失败：' + error.message);
    }
});

// 点击模态框外部关闭
document.getElementById('create-user-modal').addEventListener('click', (e) => {
    if (e.target === document.getElementById('create-user-modal')) {
        closeCreateUserModal();
    }
});

// 编辑用户表单提交
document.getElementById('edit-user-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const userId = document.getElementById('edit-user-id').value;
    const username = document.getElementById('edit-username').value.trim();
    const password = document.getElementById('edit-password').value;
    const email = document.getElementById('edit-email').value.trim();
    const role = parseInt(document.getElementById('edit-role').value);
    const status = parseInt(document.getElementById('edit-status').value);
    
    const data = { username, email, role, status };
    if (password) {
        data.password = password;
    }
    
    try {
        const result = await updateUser(userId, data);
        
        if (result.code === 0) {
            closeEditUserModal();
            loadUsers();
            alert('用户更新成功');
        } else {
            alert(result.message || '更新失败');
        }
    } catch (error) {
        console.error('更新用户失败:', error);
        alert('更新用户失败：' + error.message);
    }
});

// 点击模态框外部关闭
document.getElementById('edit-user-modal').addEventListener('click', (e) => {
    if (e.target === document.getElementById('edit-user-modal')) {
        closeEditUserModal();
    }
});

// 截断文本
function truncateText(text, maxLength) {
    if (!text) return '-';
    return text.length > maxLength ? text.substring(0, maxLength) + '...' : text;
}



// ==================== 事件监听 ====================

// 表单提交处理
document.getElementById('shorten-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const longURL = document.getElementById('long-url-input').value.trim();
    const expireDays = parseInt(document.getElementById('expire-days').value);
    const linkName = document.getElementById('link-name').value.trim();
    const linkRemark = document.getElementById('link-remark').value.trim();
    
    if (!longURL) {
        alert('请输入需要缩短的链接');
        return;
    }
    
    try {
        const result = await createShortURL(longURL, expireDays, '', linkName, linkRemark);
        if (result.code === 0) {
            showResult(result.data.short_code);
            updateStatsCards();
        } else {
            alert(result.message || '生成失败');
        }
    } catch (error) {
        console.error('生成短链失败:', error);
        alert('生成失败，请检查网络连接或重试');
    }
});

// 复制按钮处理
document.getElementById('copy-btn').addEventListener('click', () => {
    const shortURL = document.getElementById('short-url-result').value;
    copyToClipboard(shortURL);
});

// 结果输入框点击复制
document.getElementById('short-url-result').addEventListener('click', function() {
    copyToClipboard(this.value);
});

// 登录表单处理
document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const username = document.getElementById('login-username').value.trim();
    const password = document.getElementById('login-password').value;
    
    if (!username || !password) {
        alert('请输入用户名和密码');
        return;
    }
    
    try {
        const result = await login(username, password);
        if (result.code === 0) {
            authToken = result.data.token;
            localStorage.setItem('token', authToken);
            currentUser = result.data.user;
            
            updateNavbar();
            showPage('home');
            alert('登录成功');
        } else {
            alert(result.message || '登录失败');
        }
    } catch (error) {
        console.error('登录失败:', error);
        alert(error.message);
    }
});

// 修改密码表单处理
document.getElementById('change-password-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const oldPassword = document.getElementById('old-password').value;
    const newPassword = document.getElementById('new-password').value;
    
    if (!oldPassword || !newPassword) {
        alert('请输入旧密码和新密码');
        return;
    }
    
    if (newPassword.length < 6) {
        alert('新密码长度不能少于6位');
        return;
    }
    
    try {
        const result = await changePassword(oldPassword, newPassword);
        if (result.code === 0) {
            alert('密码修改成功，请重新登录');
            logout();
        } else {
            alert(result.message || '修改失败');
        }
    } catch (error) {
        console.error('修改密码失败:', error);
        alert(error.message);
    }
});