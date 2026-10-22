import english from '~/locales/en.json'
import french from '~/locales/fr.json'

const dictionaries = { en: english, fr: french }
const storageKey = 'disaster-network-language'

function initialLocale() {
  if (import.meta.client) {
    try { return localStorage.getItem(storageKey) === 'fr' ? 'fr' : 'en' }
    catch { /* Language selection still works when browser storage is unavailable. */ }
  }
  return 'en'
}

export function useAppI18n() {
  const locale = useState('disaster-network-language', initialLocale)

  function setLocale(value) {
    if (!Object.hasOwn(dictionaries, value)) return
    locale.value = value
    if (import.meta.client) {
      try { localStorage.setItem(storageKey, value) }
      catch { /* Keep the selected language in memory when storage is unavailable. */ }
    }
  }

  // Look up the entire English phrase literally, including periods and punctuation.
  function t(phrase, parameters = {}) {
    if (!phrase) return ''
    const dictionary = dictionaries[locale.value] || english
    const translated = Object.hasOwn(dictionary, phrase) ? dictionary[phrase]
      : Object.hasOwn(english, phrase) ? english[phrase] : phrase
    return translated.replace(/\{(\w+)\}/g, (placeholder, key) => Object.hasOwn(parameters, key) ? String(parameters[key]) : placeholder)
  }

  function formatDate(value, options = {}) {
    const date = typeof value === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(value)
      ? new Date(`${value}T12:00:00Z`) : new Date(value)
    return new Intl.DateTimeFormat(locale.value === 'fr' ? 'fr-FR' : 'en-GB', {
      day: 'numeric', month: 'long', year: 'numeric', ...options, timeZone: 'Europe/Paris'
    }).format(date)
  }

  function formatTime(value) {
    return new Intl.DateTimeFormat(locale.value === 'fr' ? 'fr-FR' : 'en-GB', {
      hour: '2-digit', minute: '2-digit', timeZone: 'Europe/Paris'
    }).format(new Date(value))
  }

  function formatMoney(value) {
    return new Intl.NumberFormat(locale.value === 'fr' ? 'fr-FR' : 'en-GB', { style: 'currency', currency: 'EUR', maximumFractionDigits: 0 }).format(value)
  }

  return { locale, setLocale, t, formatDate, formatTime, formatMoney }
}
