module Something = struct
  type t = Cmt of string | Cmti of string

  let to_string = function
    | Cmt s -> s
    | Cmti s -> s
  ;;
end
