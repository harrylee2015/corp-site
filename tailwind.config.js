/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/templates/**/*.html"],
  theme: {
    extend: {
      fontFamily: {
        sans: ['"Noto Sans SC"', 'system-ui', '-apple-system', '"PingFang SC"', '"Microsoft YaHei"', 'sans-serif'],
      },
      colors: {
        brand: { DEFAULT: '#2563eb', light: '#3b82f6', dark: '#1d4ed8' },
        energy: { DEFAULT: '#10b981', light: '#34d399' },
        finance: { DEFAULT: '#f59e0b', light: '#fbbf24' },
        lease: { DEFAULT: '#8b5cf6', light: '#a78bfa' },
      },
    },
  },
  plugins: [],
};
