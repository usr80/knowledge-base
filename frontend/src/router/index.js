import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue')
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/Register.vue')
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/views/Layout.vue'),
    redirect: { name: 'Documents' },
    children: [
      {
        path: 'documents',
        name: 'Documents',
        component: () => import('@/views/Documents.vue')
      },
      {
        path: 'documents/:id',
        name: 'DocumentDetail',
        component: () => import('@/views/DocumentDetail.vue')
      },
      {
        path: 'document/edit/:id?',
        name: 'DocumentEdit',
        component: () => import('@/views/DocumentEdit.vue')
      },
      {
        path: 'categories',
        name: 'Categories',
        component: () => import('@/views/Categories.vue')
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/Profile.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const isAuthPage = to.name === 'Login' || to.name === 'Register'

  if (!isAuthPage) {
    if (!token) {
      next({ name: 'Login' })
    } else {
      next()
    }
  } else {
    if (token) {
      next({ name: 'Documents' })
    } else {
      next()
    }
  }
})

export default router
