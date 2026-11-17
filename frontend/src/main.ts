import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'
import { useToastStore } from './stores/toast'
import { installGlobalErrorHandlers } from './utils/observability'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

const toast = useToastStore(pinia)
installGlobalErrorHandlers(app, toast.error)

app.mount('#app')
