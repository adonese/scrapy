/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./web/ui/**/*.templ",
    "./web/ui/**/*_templ.go",
    "./web/static/src/**/*.{js,ts}"
  ],
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'sans-serif'],
      },
      colors: {
        primary: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
          950: '#172554',
        },
      },
      maxWidth: {
        // Custom container width matching original design (1200px)
        'container': '1200px',
      },
      letterSpacing: {
        // Custom letter spacing for eyebrow text
        'widest-plus': '0.35em',
      },
      boxShadow: {
        // Custom shadows matching the original design
        'panel': '0 20px 45px -30px rgba(15, 23, 42, 0.4)',
        'button': '0 10px 20px -10px rgba(15, 23, 42, 0.6)',
        'button-hover': '0 14px 20px -12px rgba(15, 23, 42, 0.65)',
      },
      transitionDuration: {
        '120': '120ms',
      },
    },
  },
  plugins: [],
}
