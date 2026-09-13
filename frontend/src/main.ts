import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { autoAnimatePlugin } from '@formkit/auto-animate/vue'
import './style.css'
import App from './App.vue'

const app = createApp(App)

app.use(createPinia())
app.use(autoAnimatePlugin)
app.mount('#app')