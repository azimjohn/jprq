/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./**/*.{html,js}"],
  theme: {
    extend: {
      colors: {
        jprq: {
          bg: '#f5f8ff',
          black: {
            40: '',
          }
        },
      },
    },
  },
  plugins: [],
}