# 短链权限校验与创建人关联实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现短链有效期权限校验（游客限一周）和短链创建人关联功能，普通用户可查看自己创建的短链，管理员可查看所有短链及创建人。

**Architecture:** 在现有分层架构上渐进扩展，前端动态渲染有效期选项，后端 Handler 获取 JWT userID 并校验权限，Repository JOIN User 表获取创建人信息。

**Tech Stack:** Go + Gin + GORM + SQLite (glebarez/sqlite) + HTML/JS

---

## 文件结构

| 文件 | 操作 | 职责 |
|------|------|------|
| `internal/model/shorturl.go` | Modify | 添加 CreatedBy 字段存储创建用户名 |
| `internal/handler/shorturl_handler.go` | Modify | CreateShortURL 获取 userID、校验有效期、添加用户短链列表接口 |
| `internal/service/shorturl_service.go` | Modify | CreateShortURL 接收 userID、ListShortURLs 返回创建人 |
| `internal/repository/shorturl_repo.go` | Modify | List 方法 JOIN User 表获取 username |
| `cmd/main.go` | Modify | 添加用户短链列表路由、CreateShortURL 支持 JWT可选 |
| `static/index.html` | Modify | 有效期 select 动态渲染、用户中心添加短链管理 |
| `static/js/app.js` | Modify | 前端权限校验、用户短链列表渲染 |

---

### Task 1: 模型层添加创建人字段

**Files:**
- Modify: `internal/model/shorturl.go`

- [ ] **Step 1: 添加 CreatedBy 字段**

```go
type ShortURL struct {
	ID         uint      `gorm:"primary_key" json:"id"`
	ShortCode  string    `gorm:"unique;not null;size:12" json:"short_code"`
	Name       string    `gorm:"size:100" json:"name"`
	Remark     string    `gorm:"size:500" json:"remark"`
	LongURL    string    `gorm:"not null;size:2048" json:"long_url"`
	UserID     uint      `gorm:"default:0" json:"user_id"`
	CreatedBy  string    `gorm:"size:50" json:"created_by"`  // 新增：创建用户名
	VisitCount uint      `gorm:"default:0" json:"total_visits"`
	ExpireAt   *time.Time `gorm:"null" json:"expire_at"`
	Status     int       `gorm:"default:1" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
```

- [ ] **Step 2: 提交代码**

```bash
git add internal/model/shorturl.go
git commit -m "feat: add CreatedBy field to ShortURL model"
```

---

### Task 2: Service 层修改创建短链逻辑

**Files:**
- Modify: `internal/service/shorturl_service.go`

- [ ] **Step 1: 修改 CreateShortURL 接口签名**

```go
// 修改接口定义（约第16行）
type ShortURLService interface {
	CreateShortURL(longURL string, expireDays int, userID uint, username string, name, remark string) (*model.ShortURL, error)
	// ... 其他方法保持不变
}
```

- [ ] **Step 2: 修改 CreateShortURL 实现**

```go
func (s *shortURLService) CreateShortURL(longURL string, expireDays int, userID uint, username string, name, remark string) (*model.ShortURL, error) {
	if !util.IsValidURL(longURL) {
		return nil, ErrInvalidURL
	}

	// 检查是否已存在相同的长链接
	existing, err := s.repo.FindByLongURL(longURL)
	if err == nil && existing != nil && existing.Status == model.StatusActive {
		return existing, nil
	}

	shortCode := util.GenerateShortCode()

	now := time.Now()
	var expireAt *time.Time
	if expireDays > 0 {
		expire := now.Add(time.Duration(expireDays) * 24 * time.Hour)
		expireAt = &expire
	}

	// 创建人用户名：登录用户用 username，游客用 "guest"
	createdBy := username
	if createdBy == "" {
		createdBy = "guest"
	}

	shortURL := &model.ShortURL{
		ShortCode: shortCode,
		Name:      name,
		Remark:    remark,
		LongURL:   longURL,
		UserID:    userID,
		CreatedBy: createdBy,
		ExpireAt:  expireAt,
		Status:    model.StatusActive,
	}

	if err := s.repo.Create(shortURL); err != nil {
		return nil, err
	}

	return shortURL, nil
}
```

- [ ] **Step 3: 提交代码**

```bash
git add internal/service/shorturl_service.go
git commit -m "feat: update CreateShortURL to accept username"
```

---

### Task 3: Repository 层 JOIN User 表获取创建人

**Files:**
- Modify: `internal/repository/shorturl_repo.go`

- [ ] **Step 1: 修改 List 方法支持获取创建人**

```go
// 新增结构体用于 JOIN 查询结果
type ShortURLWithCreator struct {
	model.ShortURL
	CreatorName string `json:"creator_name"`
}

// 修改 List 方法
func (r *shortURLRepository) List(userID uint, page, limit int) ([]model.ShortURL, error) {
	var shortURLs []model.ShortURL
	offset := (page - 1) * limit

	if userID == 0 {
		// 管理员：查询所有短链
		err := r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&shortURLs).Error
		if err != nil {
			return nil, err
		}
	} else {
		// 普通用户：只查自己的
		err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&shortURLs).Error
		if err != nil {
			return nil, err
		}
	}

	return shortURLs, nil
}

// 修改 Count 方法
func (r *shortURLRepository) Count(userID uint) (int, error) {
	var count int
	if userID == 0 {
		err := r.db.Model(&model.ShortURL{}).Count(&count).Error
		if err != nil {
			return 0, err
		}
	} else {
		err := r.db.Model(&model.ShortURL{}).Where("user_id = ?", userID).Count(&count).Error
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}
```

- [ ] **Step 2: 提交代码**

```bash
git add internal/repository/shorturl_repo.go
git commit -m "feat: update List/Count to support userID filter for all/user"
```

---

### Task 4: Handler 层添加权限校验和用户短链接口

**Files:**
- Modify: `internal/handler/shorturl_handler.go`

- [ ] **Step 1: 修改 CreateShortURL Handler 获取 userID 和校验有效期**

```go
func (h *ShortURLHandler) CreateShortURL(c *gin.Context) {
	var req struct {
		LongURL    string `json:"long_url" binding:"required"`
		ExpireDays int    `json:"expire_days"`
		Name       string `json:"name"`
		Remark     string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 获取当前用户信息（游客为 nil）
	var userID uint = 0
	var username string = ""
	
	userIDValue, exists := c.Get("user_id")
	if exists {
		userID = userIDValue.(uint)
	}
	
	usernameValue, exists := c.Get("username")
	if exists {
		username = usernameValue.(string)
	}

	// 游客权限校验：只能创建一周内的短链
	if userID == 0 && req.ExpireDays > 7 && req.ExpireDays != 0 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "游客只能创建有效期一周内的短链"})
		return
	}

	shortURL, err := h.service.CreateShortURL(req.LongURL, req.ExpireDays, userID, username, req.Name, req.Remark)
	if err != nil {
		if err == service.ErrInvalidURL {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的URL"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	host := c.Request.Host
	shortURLStr := "http://" + host + "/" + shortURL.ShortCode

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"short_url":  shortURLStr,
			"short_code": shortURL.ShortCode,
			"expire_at":  shortURL.ExpireAt,
			"created_by": shortURL.CreatedBy,
		},
	})
}
```

- [ ] **Step 2: 添加用户短链列表接口**

```go
// ListUserShortURLs 用户查看自己的短链列表
func (h *ShortURLHandler) ListUserShortURLs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	shortURLs, total, err := h.service.ListShortURLs(userID.(uint), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"items": shortURLs,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// ListAdminShortURLs 管理员查看所有短链列表
func (h *ShortURLHandler) ListAdminShortURLs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	// userID = 0 表示查询所有
	shortURLs, total, err := h.service.ListShortURLs(0, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"items": shortURLs,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}
```

- [ ] **Step 3: 提交代码**

```bash
git add internal/handler/shorturl_handler.go
git commit -m "feat: add permission check and user short URL list handler"
```

---

### Task 5: JWT Middleware 添加 username 信息

**Files:**
- Modify: `internal/middleware/jwt.go`

- [ ] **Step 1: 在 JWT middleware 中添加 username 到 context**

```go
func JWTMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 不强制要求登录，继续执行（游客模式）
			c.Set("user_id", uint(0))
			c.Set("username", "")
			c.Next()
			return
		}

		// ... JWT 解析逻辑
		// 成功后设置 user_id 和 username
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
```

- [ ] **Step 2: 提交代码**

```bash
git add internal/middleware/jwt.go
git commit -m "feat: add username to JWT context for short URL creation"
```

---

### Task 6: 路由配置修改

**Files:**
- Modify: `cmd/main.go`

- [ ] **Step 1: 修改 CreateShortURL 路由支持可选 JWT**

```go
// 公开接口改为支持可选 JWT（游客可创建，登录用户也可创建）
api.POST("/short-url", middleware.OptionalJWTMiddleware(cfg.JWTSecret), shortURLHandler.CreateShortURL)
```

- [ ] **Step 2: 添加用户短链列表路由**

```go
// 需要登录的接口
userGroup := api.Group("/user")
userGroup.Use(middleware.JWTMiddleware(cfg.JWTSecret))
{
	userGroup.GET("/profile", userHandler.GetProfile)
	userGroup.PUT("/password", userHandler.ChangePassword)
	userGroup.PUT("/profile", userHandler.UpdateProfile)
	// 新增：用户短链列表
	userGroup.GET("/short-url/list", shortURLHandler.ListUserShortURLs)
}

// 管理员短链管理改为使用 ListAdminShortURLs
adminGroup.GET("/short-url/list", shortURLHandler.ListAdminShortURLs)
```

- [ ] **Step 3: 添加 OptionalJWTMiddleware 函数**

在 `internal/middleware/jwt.go` 中添加：

```go
// OptionalJWTMiddleware 可选 JWT 认证（游客和登录用户都可访问）
func OptionalJWTMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 游客模式
			c.Set("user_id", uint(0))
			c.Set("username", "")
			c.Next()
			return
		}

		// 尝试解析 JWT
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ParseToken(tokenString, jwtSecret)
		if err != nil {
			// JWT 无效，仍然允许访问（游客模式）
			c.Set("user_id", uint(0))
			c.Set("username", "")
			c.Next()
			return
		}

		// JWT 有效
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
```

- [ ] **Step 4: 提交代码**

```bash
git add cmd/main.go internal/middleware/jwt.go
git commit -m "feat: add optional JWT middleware and user short URL list route"
```

---

### Task 7: 前端有效期动态渲染

**Files:**
- Modify: `static/js/app.js`

- [ ] **Step 1: 添加有效期动态渲染函数**

```javascript
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

// 在 checkLoginStatus 成功后调用
async function checkLoginStatus() {
    if (authToken) {
        try {
            const result = await getProfile();
            if (result.code === 0) {
                currentUser = result.data;
                updateNavbar();
                updateExpireOptions();  // 新增
            }
        } catch (error) {
            console.error('检查登录状态失败:', error);
            logout();
        }
    } else {
        // 游客状态也要渲染选项
        updateExpireOptions();
    }
}
```

- [ ] **Step 2: 提交代码**

```bash
git add static/js/app.js
git commit -m "feat: dynamic expire options based on login status"
```

---

### Task 8: 前端添加用户短链管理功能

**Files:**
- Modify: `static/index.html`
- Modify: `static/js/app.js`

- [ ] **Step 1: 在用户中心添加短链管理标签页**

在 `static/index.html` 的 `page-profile` 区域添加：

```html
<!-- 用户中心 -->
<div id="page-profile" class="page hidden">
    <div class="profile-container">
        <h2>用户中心</h2>
        
        <!-- 标签页 -->
        <div class="profile-tabs">
            <button class="tab-btn active" onclick="showProfileTab('info')">个人信息</button>
            <button class="tab-btn" onclick="showProfileTab('mylinks')">我的短链</button>
        </div>
        
        <!-- 个人信息 -->
        <div id="tab-info" class="profile-tab active">
            <div class="profile-section">
                <h3>个人信息</h3>
                <div class="info-grid">
                    <div class="info-item">
                        <span class="info-label">用户名</span>
                        <span class="info-value" id="profile-username">-</span>
                    </div>
                    <div class="info-item">
                        <span class="info-label">邮箱</span>
                        <span class="info-value" id="profile-email">-</span>
                    </div>
                    <div class="info-item">
                        <span class="info-label">角色</span>
                        <span class="info-value" id="profile-role">-</span>
                    </div>
                </div>
            </div>

            <!-- 修改密码 -->
            <div class="profile-section">
                <h3>修改密码</h3>
                <form id="change-password-form" class="change-password-form">
                    <div class="form-group">
                        <label for="old-password">旧密码</label>
                        <input type="password" id="old-password" placeholder="请输入旧密码" required>
                    </div>
                    <div class="form-group">
                        <label for="new-password">新密码</label>
                        <input type="password" id="new-password" placeholder="请输入新密码" required>
                    </div>
                    <button type="submit" class="submit-btn">修改密码</button>
                </form>
            </div>
        </div>
        
        <!-- 我的短链 -->
        <div id="tab-mylinks" class="profile-tab hidden">
            <div class="profile-section">
                <h3>我创建的短链</h3>
                <div class="admin-actions">
                    <button class="action-btn" onclick="loadMyShortURLs()">刷新列表</button>
                </div>
                <table class="admin-table">
                    <thead>
                        <tr>
                            <th>短码</th>
                            <th>名称</th>
                            <th>长链接</th>
                            <th>访问量</th>
                            <th>剩余天数</th>
                            <th>状态</th>
                        </tr>
                    </thead>
                    <tbody id="mylinks-table-body">
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</div>
```

- [ ] **Step 2: 添加 CSS 样式支持 profile-tabs**

在 `static/css/main.css` 中添加：

```css
/* 用户中心标签页 */
.profile-tabs {
    display: flex;
    gap: 10px;
    margin-bottom: 20px;
}

.profile-tabs .tab-btn {
    padding: 10px 20px;
    border: none;
    background: #f0f0f0;
    cursor: pointer;
    border-radius: 5px;
}

.profile-tabs .tab-btn.active {
    background: #1890ff;
    color: white;
}

.profile-tab {
    display: block;
}

.profile-tab.hidden {
    display: none;
}
```

- [ ] **Step 3: 在 app.js 添加用户短链管理函数**

```javascript
// 显示用户中心标签页
function showProfileTab(tabName) {
    document.querySelectorAll('.profile-tab').forEach(tab => {
        tab.classList.remove('active');
        tab.classList.add('hidden');
    });
    document.querySelectorAll('.profile-tabs .tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    const targetTab = document.getElementById(`tab-${tabName}`);
    targetTab.classList.remove('hidden');
    targetTab.classList.add('active');
    document.querySelector(`[onclick="showProfileTab('${tabName}')"]`).classList.add('active');
    
    if (tabName === 'mylinks') {
        loadMyShortURLs();
    }
}

// 加载用户自己的短链列表
async function loadMyShortURLs() {
    try {
        const response = await fetch(`${API_BASE_URL}/user/short-url/list`, {
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        if (!response.ok) {
            throw new Error('获取短链列表失败');
        }
        
        const result = await response.json();
        if (result.code === 0) {
            renderMyShortURLTable(result.data.items);
        }
    } catch (error) {
        console.error('加载短链列表失败:', error);
        alert('加载短链列表失败');
    }
}

// 渲染用户短链表格（无创建人列）
function renderMyShortURLTable(items) {
    const tbody = document.getElementById('mylinks-table-body');
    tbody.innerHTML = '';
    
    if (!items || items.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;">暂无数据</td></tr>';
        return;
    }
    
    items.forEach(item => {
        const shortURL = `${window.location.protocol}//${window.location.host}/${item.short_code}`;
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

// 辅助函数：截断文本
function truncateText(text, maxLength) {
    if (!text) return '';
    return text.length > maxLength ? text.substring(0, maxLength) + '...' : text;
}
```

- [ ] **Step 4: 提交代码**

```bash
git add static/index.html static/css/main.css static/js/app.js
git commit -m "feat: add user short URL management in profile page"
```

---

### Task 9: 管理员短链列表添加创建人列

**Files:**
- Modify: `static/js/app.js`

- [ ] **Step 1: 修改管理员短链表格渲染添加创建人列**

```javascript
// 渲染短链表格（管理员版本，含创建人）
function renderShortURLTable(items) {
    const tbody = document.getElementById('shorturls-table-body');
    tbody.innerHTML = '';
    
    if (!items || items.length === 0) {
        tbody.innerHTML = '<tr><td colspan="9" style="text-align:center;">暂无数据</td></tr>';
        return;
    }
    
    items.forEach(item => {
        const shortURL = `${window.location.protocol}//${window.location.host}/${item.short_code}`;
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
```

- [ ] **Step 2: 修改 HTML 表格添加创建人列头**

在 `static/index.html` 的管理员短链表格中：

```html
<thead>
    <tr>
        <th>短码</th>
        <th>名称</th>
        <th>备注</th>
        <th>长链接</th>
        <th>访问量</th>
        <th>剩余天数</th>
        <th>创建人</th>
        <th>状态</th>
        <th>操作</th>
    </tr>
</thead>
```

- [ ] **Step 3: 提交代码**

```bash
git add static/index.html static/js/app.js
git commit -m "feat: add created_by column to admin short URL table"
```

---

### Task 10: 测试验证

- [ ] **Step 1: 启动项目**

```bash
go run cmd/main.go
```

- [ ] **Step 2: 验证游客权限**

1. 未登录状态下，检查有效期下拉只有 1天、3天、1周
2. 创建短链选择超过 7天，应返回 403 错误
3. 创建短链成功后，created_by 应为 "guest"

- [ ] **Step 3: 验证登录用户权限**

1. 登录普通用户
2. 检查有效期下拉显示全部选项（含永久）
3. 创建短链成功后，created_by 应为用户名
4. 进入用户中心 → 我的短链，应只显示自己创建的短链

- [ ] **Step 4: 验证管理员权限**

1. 登录管理员账号
2. 进入管理后台 → 短链管理
3. 应显示所有短链，含创建人列
4. 创建人列显示用户名或 guest

---

### Task 11: 最终提交

- [ ] **Step 1: 运行 go mod tidy**

```bash
go mod tidy
```

- [ ] **Step 2: 提交所有修改并推送**

```bash
git add -A
git commit -m "feat: complete short URL permission and creator association"
git push
```

---

## Spec 覆盖检查

| 需求 | 任务覆盖 |
|------|----------|
| 游客只能选择有效期一周以下 | Task 4 (Handler校验) + Task 7 (前端动态渲染) |
| 登录用户可选择所有有效期含永久 | Task 7 (前端动态渲染) |
| 生成短链关联创建用户 | Task 1 (模型) + Task 2 (Service) + Task 4 (Handler) |
| 普通用户查看自己生成的短链 | Task 3 (Repository) + Task 4 (Handler) + Task 8 (前端) |
| 管理员查看所有短链含创建人 | Task 3 (Repository) + Task 9 (前端) |
| 短链管理显示创建人 | Task 9 (前端表格) |

---

**Plan complete.** Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session, batch execution with checkpoints

**Which approach?**