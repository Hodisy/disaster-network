import { createHash, timingSafeEqual } from 'node:crypto'
import { resetDemoDataset } from '../repositories/demoDatabase.js'

export class DemoResetError extends Error {
  constructor(code, statusCode) {
    super(code)
    this.code = code
    this.statusCode = statusCode
  }
}

export async function resetDemoDatabase({ config, maintenanceKey, confirmation }) {
  if (config.demoResetEnabled !== true && config.demoResetEnabled !== 'true') {
    throw new DemoResetError('RESET_DISABLED', 403)
  }
  const expected = config.demoResetToken
  if (typeof expected !== 'string' || expected.length < 32 || expected.length > 256) {
    throw new DemoResetError('RESET_NOT_CONFIGURED', 503)
  }
  if (typeof maintenanceKey !== 'string' || maintenanceKey.length > 256) {
    throw new DemoResetError('INVALID_MAINTENANCE_KEY', 403)
  }
  const digest = value => createHash('sha256').update(value).digest()
  if (!timingSafeEqual(digest(expected), digest(maintenanceKey))) {
    throw new DemoResetError('INVALID_MAINTENANCE_KEY', 403)
  }
  if (confirmation !== 'RESET_DEMO_DATABASE') {
    throw new DemoResetError('RESET_CONFIRMATION_REQUIRED', 400)
  }
  // Authentication here protects a maintenance operation, not a customer login.
  return resetDemoDataset(config)
}
