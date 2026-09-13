/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        fnos: {
          blue: '#1669ff',
          'blue-hover': '#0052e0',
          bg: '#f4f6fb',
          'bg-dark': '#12141a',
          sidebar: '#f7f9fc',
          'sidebar-dark': '#181b22',
          card: '#ffffff',
          'card-dark': '#1e212b',
        }
      }
    },
  },
  plugins: [],
}
