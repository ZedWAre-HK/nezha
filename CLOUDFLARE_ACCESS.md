# Cloudflare Access OAuth2/OIDC 登录

此版本支持将 Cloudflare Access 作为哪吒面板的 OIDC 身份提供方。管理员配置可使用 Cloudflare 返回的邮箱；数据库仍以稳定的 OIDC `sub` 作为用户标识，以兼容已有数据。

## 1. 创建 Cloudflare Access 应用

1. 进入 Cloudflare Zero Trust 控制台的 **Access controls > Applications**。
2. 创建 **SaaS application**，协议选择 **OIDC**。
3. Redirect URL 填写 `https://<哪吒面板域名>/oauth2/callback`。
4. Scopes 只启用 `openid`、`email`、`profile`。此登录流程不请求 `groups`，以兼容未提供组 claim 的身份提供程序。
5. 创建允许登录面板的 Access policy。
6. 保存以下值：Client ID、Client secret 和团队域名，例如 `https://my-team.cloudflareaccess.com`。

## 2. 配置哪吒面板

编辑面板数据目录中的 `config.yaml`：

```yaml
oauth2:
  type: cloudflare
  admin: admin@example.com
  clientid: <Cloudflare Access Client ID>
  clientsecret: <Cloudflare Access Client secret>
  endpoint: https://<team-name>.cloudflareaccess.com
```

多个管理员邮箱使用半角逗号分隔。`endpoint` 只填写团队域名，不要追加 `/cdn-cgi/...` 路径。

重启面板后访问 `/login` 测试。Cloudflare 中登记的 Redirect URL 必须与浏览器实际访问的协议、域名和 `/oauth2/callback` 路径完全一致。面板位于反向代理后时，请确保代理传递 `Host` 和 `X-Forwarded-Proto` 请求头。
