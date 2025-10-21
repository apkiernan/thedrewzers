/** @type {import('tailwindcss').Config} */
export default {
  content: ["./internal/views/*.templ"], // this is where our templates are located
  theme: {
    extend: {
      fontFamily: {
        'libre-bodoni': ['"Libre Bodoni"', 'serif'],
      },
    },
  },
  plugins: [],
};
