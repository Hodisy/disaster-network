import { hotlineRequest } from './useHotlineClaim.js'

export function useHotlineAccess() {
  const access = useState('hotline-access', () => ({ authorized: false, checking: false, error: '' }))

  async function restoreAccess() {
    access.value.checking = true
    try {
      await hotlineRequest('access')
      access.value.authorized = true
      access.value.error = ''
    } catch (error) {
      access.value.authorized = false
      access.value.error = error.code || 'HOTLINE_FAILED'
    } finally { access.value.checking = false }
  }

  return { access, restoreAccess }
}
