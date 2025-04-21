<script setup lang="ts">
import { ref, nextTick } from 'vue'
import MarkdownIt from 'markdown-it'

const input = ref('')
const messages = ref<{ role: string; content: string }[]>([])
const loading = ref(false)
const md = new MarkdownIt()
const chatContainer = ref<HTMLDivElement | null>(null)

async function send() {
  if (!input.value.trim()) return
  loading.value = true
  const question = input.value
  input.value = ''
  messages.value.push({ role: 'user', content: question })

  // 先插入一个 assistant 消息用于流式追加
  const assistantMsg = { role: 'assistant', content: '' }
  messages.value.push(assistantMsg)

  // fetch + ReadableStream 处理 SSE
  const resp = await fetch('http://localhost:7080/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ prompt: question }),
  })
  if (!resp.body) {
    loading.value = false
    return
  }
  const reader = resp.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let done = false
  while (!done) {
    const { value, done: doneReading } = await reader.read()
    if (value) {
      const chunk = decoder.decode(value)
      // SSE格式处理
      chunk.split('\n\n').forEach(line => {
        if (line.startsWith('data: ')) {
          const data = line.slice(6)
          if (data === '[DONE]') {
            loading.value = false
          } else {
            assistantMsg.content += data
            // 强制更新视图以实现流式输出
            messages.value = [...messages.value]
          }
        }
      })
      await nextTick()
      if (chatContainer.value) {
        chatContainer.value.scrollTop = chatContainer.value.scrollHeight
      }
    }
    done = doneReading
  }
}
</script>

<template>
  <main class="flex flex-col items-center min-h-screen bg-base-200">
    <h1 class="text-2xl font-bold my-4">AI 对话</h1>
    <div class="w-full max-w-2xl flex flex-col gap-4">
      <div ref="chatContainer" class="bg-base-100 rounded-box p-4 h-[60vh] overflow-y-auto shadow"
        style="scroll-behavior:smooth;">
        <div v-for="(msg, i) in messages" :key="i" class="mb-4">
          <div class="chat" :class="msg.role === 'user' ? 'chat-end' : 'chat-start'">
            <div class="chat-header mb-1 text-xs text-gray-400">
              {{ msg.role === 'user' ? '你' : 'AI' }}
            </div>
            <div class="chat-bubble whitespace-pre-line max-w-full prose prose-sm" v-html="md.render(msg.content)" />
          </div>
        </div>
        <div
          v-if="loading && messages[messages.length - 1]?.role === 'assistant' && !messages[messages.length - 1]?.content"
          class="chat chat-start">
          <div class="chat-header mb-1 text-xs text-gray-400">AI</div>
          <div class="chat-bubble loading">AI 正在思考...</div>
        </div>
      </div>
      <form class="flex gap-2" @submit.prevent="send">
        <input v-model="input" class="input input-bordered flex-1" placeholder="请输入你的问题..." :disabled="loading"
          autocomplete="off" />
        <button class="btn btn-primary" :disabled="loading || !input.trim()">发送</button>
      </form>
    </div>
  </main>
</template>