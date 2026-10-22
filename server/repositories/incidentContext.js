import context from '../../docs/scenarios/hotline-incident-context.json' with { type: 'json' }

// Versioned local scenario evidence, bundled with Nitro. No simulated network service.
export function readIncidentContext() {
  return structuredClone(context)
}
