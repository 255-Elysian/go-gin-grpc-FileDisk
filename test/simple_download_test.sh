#!/bin/bash

# 简单的七牛云文件下载测试脚本

# 配置参数
FILE_ID="41"  # 请替换为实际的文件ID
TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTMyOTUxOTksIm5iZiI6MTc1MzIwODc5OSwiaWF0IjoxNzUzMjA4Nzk5LCJ1c2VyX2lkIjoyfQ.BTbNz8iyciZiZh3kGIK--A-fz2AFa58p93NZBI521G_m8fafK2dlVJ4hXrzss--ApRoOqy28pbZbXJpe7Un-n80xmTfJur8nArh-_GfInEgjU25YDNOR2a9_7CveGKO1DFtZZrhL0lm8Qwnjo2SjwwKAmeFshyiG1fQ0uWyvit2hiE1abwdIwK_pOtrlPLupnwlPMnCzSfJ-9COJvCQx2FtwC1x933KkUSKnlYwWfy-jPBkJxilp1nCANEDbFWrju5R3A7qDprnvxTLEFuHnKuX6kjA1m250QznGHSj4PJRNBYyDQHAiR9VP9kVshPm02bMSm3DOgIT_to0mqLcJ_g"  # 请替换为有效的JWT token
GATEWAY_URL="http://localhost:4000"

echo "=== 简单七牛云文件下载测试 ==="
echo "文件ID: $FILE_ID"
echo "Gateway地址: $GATEWAY_URL"
echo ""

# 获取下载链接
echo "🔍 正在获取下载链接..."
RESPONSE=$(curl -s -X GET \
  "$GATEWAY_URL/api/v1/qiniu_file_download?file_id=$FILE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json")

echo "完整API响应:"
echo "$RESPONSE"
echo ""

# 检查响应是否成功
if echo "$RESPONSE" | grep -q '"status":200'; then
    echo "✅ API调用成功"
    
    # 使用更简单的方法提取URL和文件名
    # 提取download_url
    DOWNLOAD_URL=$(echo "$RESPONSE" | sed -n 's/.*"download_url":"\([^"]*\)".*/\1/p')
    # 提取file_name
    FILENAME=$(echo "$RESPONSE" | sed -n 's/.*"file_name":"\([^"]*\)".*/\1/p')
    
    echo "提取结果:"
    echo "文件名: $FILENAME"
    echo "下载URL: $DOWNLOAD_URL"
    echo ""
    
    if [ -n "$DOWNLOAD_URL" ] && [ "$DOWNLOAD_URL" != "null" ]; then
        echo "✅ 成功提取下载链接"
        echo ""
        echo "=== 可以使用以下方式下载文件 ==="
        echo ""
        echo "1. 浏览器访问:"
        echo "   $DOWNLOAD_URL"
        echo ""
        echo "2. curl命令下载:"
        echo "   curl -L -o \"$FILENAME\" \"$DOWNLOAD_URL\""
        echo ""
        echo "3. wget命令下载:"
        echo "   wget -O \"$FILENAME\" \"$DOWNLOAD_URL\""
        echo ""
        
        # 询问是否立即下载
        read -p "是否立即使用curl下载文件? (y/n): " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo "⬇️ 正在下载文件..."
            curl -L -o "$FILENAME" "$DOWNLOAD_URL"
            
            if [ $? -eq 0 ] && [ -f "$FILENAME" ]; then
                FILE_SIZE=$(stat -c%s "$FILENAME" 2>/dev/null || stat -f%z "$FILENAME" 2>/dev/null)
                echo "✅ 文件下载成功"
                echo "文件名: $FILENAME"
                echo "文件大小: $FILE_SIZE bytes"
                echo "文件路径: $(pwd)/$FILENAME"
            else
                echo "❌ 文件下载失败"
            fi
        fi
    else
        echo "❌ 无法提取下载URL"
        echo "可能的原因:"
        echo "1. 文件不存在"
        echo "2. 没有访问权限"
        echo "3. 文件不是七牛云存储文件"
    fi
else
    echo "❌ API调用失败"
    echo "错误信息: $RESPONSE"
    echo ""
    echo "可能的原因:"
    echo "1. JWT token无效或过期"
    echo "2. 文件ID不存在"
    echo "3. Gateway服务未启动"
    echo "4. 网络连接问题"
fi

echo ""
echo "=== 测试完成 ==="
