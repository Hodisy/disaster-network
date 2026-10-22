import { hotlineRequest } from './useHotlineClaim.js'

export function useHotlineAccess() {
  const access = useState('hotline-access', () => ({ authorized: false, checking: false, error: '' }))
  const accessKey = ref('')

  async function restoreAccess() {
    access.value.checking = true
    try {
      await hotlineRequest('access')
      access.value.authorized = true
      access.value.error = ''
    } catch {
      access.value.authorized = false
    } finally { access.value.checking = false }
  }

  async function unlock() {
    if (access.value.checking || !accessKey.value.trim()) return
    access.value.checking = true
    access.value.error = ''
    try {
      await hotlineRequest('authorize', { token: accessKey.value })
      access.value.authorized = true
    } catch (error) {
      access.value.authorized = false
      access.value.error = error.code || 'HOTLINE_FAILED'
    } finally {
      accessKey.value = ''
      access.value.checking = false
    }
  }

  return { access, accessKey, restoreAccess, unlock }
}
