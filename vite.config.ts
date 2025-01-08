import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
    plugins: [react()],
    server: {
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true, // Ensures the Host header matches the backend
                secure: false, // Set to false if you're using HTTP
            },
        },
    },
});
