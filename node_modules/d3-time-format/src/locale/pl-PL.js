import locale from "../locale";

export default locale({
  dateTime: "%A, %e %B %Y, %X",
  date: "%d/%m/%Y",
  time: "%H:%M:%S",
  periods: ["AM", "PM"], // unused
  days: ["Niedziela", "Poniedziałek", "Wtorek", "Środa", "Czwartek", "Piątek", "Sobota"],
  shortDays: ["Niedz.", "Pon.", "Wt.", "Śr.", "Czw.", "Pt.", "Sob."],
  months: ["Styczeń", "Luty", "Marzec", "Kwiecień", "Maj", "Czerwiec", "Lipiec", "Sierpień", "Wrzesień", "Październik", "Listopad", "Grudzień"],
  shortMonths: ["Stycz.", "Luty", "Marz.", "Kwie.", "Maj", "Czerw.", "Lipc.", "Sierp.", "Wrz.", "Paźdz.", "Listop.", "Grudz."]/* In Polish language abbraviated months are not commonly used so there is a dispute about the proper abbraviations. */
});
