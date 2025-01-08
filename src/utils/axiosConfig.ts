// src/utils/axiosConfig.ts
import axios from 'axios';

const instance = axios.create({
  baseURL: 'http://localhost:8080', // Update with your backend URL
});

// Optionally, add interceptors to include auth token
instance.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export default instance;
