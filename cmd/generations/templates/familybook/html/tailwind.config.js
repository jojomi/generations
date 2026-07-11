module.exports = {
  // mode: 'jit',
  purge: ["./src/**/*.html", "./src/**/*.js"],
  darkMode: "media", // or 'media' or 'class'
  theme: {
    extend: {
      screens: {
        lg: { raw: "print, (min-width: 1024px)" },
        print: { raw: "print" },
      },
      gridTemplateColumns: {
        "person-dates": "1fr max-content max-content",
        "2-1st-grow": "1fr max-content",
        "2-2nd-grow": "max-content 1fr",
        "3-1st-grow": "1fr max-content max-content",
        "3-2nd-grow": "max-content 1fr max-content",
        "4-2nd-grow": "max-content 1fr max-content max-content",
      },
      minWidth: {
        'date': '10ch',
        'date-with-symbol': '12ch',
      },
      colors: {
        "non-contained": "#6B7280",
      },
    },
  },
  variants: {
    extend: {
      display: ["group-hover", "group-focus"],
      textColor: ["visited"],
    },
  },
  plugins: [],
};
