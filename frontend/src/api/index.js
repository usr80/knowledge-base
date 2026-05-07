import request from './request'

export function login(data) {
  return request.post('/auth/login', data)
}

export function register(data) {
  return request.post('/auth/register', data)
}

export function getProfile() {
  return request.get('/user/profile')
}

export function updateProfile(data) {
  return request.put('/user/profile', data)
}

export function changePassword(data) {
  return request.put('/user/password', data)
}

export function getDocuments(params) {
  return request.get('/documents', { params })
}

export function getDocument(id) {
  return request.get(`/documents/${id}`)
}

export function createDocument(data) {
  return request.post('/documents', data)
}

export function updateDocument(id, data) {
  return request.put(`/documents/${id}`, data)
}

export function deleteDocument(id) {
  return request.delete(`/documents/${id}`)
}

export function getCategories() {
  return request.get('/categories')
}

export function createCategory(data) {
  return request.post('/categories', data)
}

export function updateCategory(id, data) {
  return request.put(`/categories/${id}`, data)
}

export function deleteCategory(id) {
  return request.delete(`/categories/${id}`)
}
