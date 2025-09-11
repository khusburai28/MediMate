/* Simple client-side i18n with JSON locale files and data attributes
   - Uses localStorage 'cura_lang'
   - Detects from navigator.language if none set
   - Replaces text for elements with data-i18n
   - Supports attribute translations via data-i18n-attr="attr:key,attr2:key2"
*/
(function() {
  const DEFAULT_LANG = 'en';
  const SUPPORTED = ['en','pa','hi'];

  function getSavedLang() {
    try { return localStorage.getItem('cura_lang'); } catch(e) { return null; }
  }

  function detectLang() {
    const saved = getSavedLang();
    if (saved && SUPPORTED.includes(saved)) return saved;
    const nav = (navigator.language || navigator.userLanguage || '').toLowerCase();
    const cand = nav.split('-')[0];
    if (SUPPORTED.includes(cand)) return cand;
    return DEFAULT_LANG;
  }

  function setLang(lang) {
    try { localStorage.setItem('cura_lang', lang); } catch(e) {}
    document.documentElement.setAttribute('lang', lang);
  }

  function fetchLocale(lang) {
    const url = `/static/locales/${lang}.json`;
    return fetch(url).then(r => {
      if (!r.ok) throw new Error(`Locale ${lang} not found`);
      return r.json();
    });
  }

  function get(obj, path) {
    return path.split('.').reduce((o,k) => (o && o[k] != null) ? o[k] : null, obj);
  }

  function applyTranslations(dict) {
    document.querySelectorAll('[data-i18n]').forEach(el => {
      const key = el.getAttribute('data-i18n');
      const val = get(dict, key);
      if (val != null) {
        if (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA') {
          el.setAttribute('value', ''); // avoid showing stale values
        }
        el.textContent = val;
      }
    });
    document.querySelectorAll('[data-i18n-attr]').forEach(el => {
      const mapping = el.getAttribute('data-i18n-attr');
      mapping.split(',').forEach(pair => {
        const [attr, key] = pair.split(':').map(s => s.trim());
        const val = get(dict, key);
        if (attr && val != null) el.setAttribute(attr, val);
      });
    });
  }

  function ensureLanguageSelector(current) {
    const existing = document.getElementById('languageSelect');
    if (existing) { existing.value = current; return existing; }
    const select = document.createElement('select');
    select.id = 'languageSelect';
    select.className = 'lang-select';
    const options = [
      ['en','English'],
      ['hi','हिन्दी'],
      ['pa','ਪੰਜਾਬੀ']
    ];
    options.forEach(([val,label]) => {
      const opt = document.createElement('option');
      opt.value = val; opt.textContent = label; select.appendChild(opt);
    });
    select.value = current;
    // Place near auth-buttons if present
    const nav = document.querySelector('.navbar .container');
    if (nav) {
      const wrapper = document.createElement('div');
      wrapper.className = 'lang-wrapper';
      wrapper.style.marginLeft = '10px';
      wrapper.appendChild(select);
      const auth = nav.querySelector('.auth-buttons');
      if (auth && auth.parentNode) auth.parentNode.insertBefore(wrapper, auth.nextSibling);
      else nav.appendChild(wrapper);
    } else {
      document.body.appendChild(select);
    }
    return select;
  }

  async function init() {
    const lang = detectLang();
    setLang(lang);
    try {
      const dict = await fetchLocale(lang);
      applyTranslations(dict);
    } catch(e) {
      if (lang !== DEFAULT_LANG) {
        try {
          const dict = await fetchLocale(DEFAULT_LANG);
          applyTranslations(dict);
        } catch(_) {}
      }
    }
    const sel = ensureLanguageSelector(lang);
    sel.addEventListener('change', async (e) => {
      const next = e.target.value;
      if (!SUPPORTED.includes(next)) return;
      setLang(next);
      try {
        const dict = await fetchLocale(next);
        applyTranslations(dict);
      } catch(_) {}
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();


