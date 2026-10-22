const errors = {
  RESET_DISABLED: 'Database reset is disabled on this server.',
  RESET_NOT_CONFIGURED: 'The server maintenance key has not been configured.',
  INVALID_MAINTENANCE_KEY: 'The maintenance key is incorrect.',
  RESET_CONFIRMATION_REQUIRED: 'Confirm the reset before continuing.',
  DATABASE_NOT_CONFIGURED: 'The database connection or certificate has not been configured.',
  DEMO_SCHEMA_MISSING: 'The demo database must be set up before it can be reset.',
  DEMO_SCHEMA_MISMATCH: 'The database does not match the expected demo schema. Nothing was reset.',
  DATABASE_RESET_FAILED: 'The database reset could not be confirmed. Check the connection before trying again.',
  INVALID_REQUEST: 'The reset request was refused.'
}

export function useDemoSettings() {
  const { reset } = useDemo()
  const pending = ref(false)
  const confirming = ref(false)
  const maintenanceKey = ref('')
  const error = ref('')
  const result = ref(null)

  function requestReset() {
    if (pending.value) return
    error.value = ''
    result.value = null
    confirming.value = true
  }

  function cancelReset() {
    if (pending.value) return
    confirming.value = false
    maintenanceKey.value = ''
    error.value = ''
  }

  async function resetDatabase() {
    if (pending.value || !confirming.value || !maintenanceKey.value.trim()) return
    pending.value = true
    error.value = ''
    try {
      const response = await $fetch('/api/demo/reset', {
        method: 'POST', retry: 0,
        headers: { 'x-demo-reset-key': maintenanceKey.value },
        body: { confirmation: 'RESET_DEMO_DATABASE' }
      })
      if (!response?.ok || !response.counts) throw new Error('RESET_FAILED')
      // Browser state is cleared only after the database confirms success.
      reset()
      result.value = response.counts
      confirming.value = false
      maintenanceKey.value = ''
    } catch (failure) {
      error.value = errors[failure?.data?.code] || errors.DATABASE_RESET_FAILED
    } finally {
      pending.value = false
    }
  }

  return { pending, confirming, maintenanceKey, error, result, requestReset, cancelReset, resetDatabase }
}
