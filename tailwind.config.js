/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/templates/**/*.html"],
  theme: {
    extend: {
      colors: {
        lotus: { DEFAULT: "#15803d", dark: "#166534" },
        wood: { DEFAULT: "#b45309", light: "#d97706" },
        slateink: "#0f172a",
        slatemist: "#f8fafc",
      },
    },
  },
  plugins: [],
};
