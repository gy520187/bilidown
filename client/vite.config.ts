import { defineConfig } from 'vite'

export default defineConfig({
    clearScreen: false,
    server: {
        proxy: {
            '/api': 'http://127.0.0.1:8098'
        },
        host: '0.0.0.0',
        allowedHosts: ['.monkeycode-ai.online']
    },
    build: {
        outDir: '../server/static',
        emptyOutDir: true,
    },
    css: {
        preprocessorOptions: {
            scss: {
                api: 'modern-compiler'
            }
        }
    }
})