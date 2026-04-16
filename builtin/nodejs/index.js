const http = require('http');
const https = require('https');
const { URL } = require('url');

/**
 * 环境变量强校验：导入期进行
 */
const TOKEN = process.env.BHPKG_NOTIFY_TOKEN;
const CHANNEL = process.env.BHPKG_NOTIFY_CHANNEL;

if (!TOKEN || !CHANNEL) {
    const missing = [];
    if (!TOKEN) missing.push("BHPKG_NOTIFY_TOKEN");
    if (!CHANNEL) missing.push("BHPKG_NOTIFY_CHANNEL");
    
    throw new Error(`缺少必要的环境变量以使用 baihu 模块: ${missing.join(", ")}。请在白虎面板的任务设置中配置这些 Key。`);
}

/**
 * 发送通知的辅助函数 (仅使用 Node.js 标准库)
 */
function notify(title, text, channelId) {
    const notifyUrl = process.env.BHPKG_NOTIFY_URL || 'http://localhost:8052/api/v1/notify/send';
    const cid = channelId || CHANNEL;

    if (!notifyUrl || !TOKEN || !cid) return;

    try {
        const parsedUrl = new URL(notifyUrl);
        const protocol = parsedUrl.protocol === 'https:' ? https : http;
        
        const data = JSON.stringify({
            channel_id: cid,
            title: title || '系统通知',
            text: text
        });

        const options = {
            hostname: parsedUrl.hostname,
            port: parsedUrl.port,
            path: parsedUrl.pathname + (parsedUrl.search || ''),
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'notify-token': TOKEN,
                'Content-Length': Buffer.byteLength(data)
            }
        };

        const req = protocol.request(options);
        req.on('error', (e) => {});
        req.write(data);
        req.end();
    } catch (e) {}
}

module.exports = { notify };
