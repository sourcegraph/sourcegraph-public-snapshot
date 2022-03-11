port module Main exposing (..)

import Browser
import Chart as C
import Chart.Attributes as CA
import Dict exposing (Dict)
import Element as E
import Element.Background as Background
import Element.Border as Border
import Element.Events
import Element.Font as F
import Element.Input as I
import Html exposing (Html, input, text)
import Html.Attributes exposing (..)
import Json.Decode as Decode exposing (Decoder, fail, field, maybe)
import Json.Decode.Pipeline
import Process
import Task
import Url.Builder
import Url.Parser exposing (..)



-- CONSTANTS


width : Int
width =
    800


debounceQueryInputMillis : Float
debounceQueryInputMillis =
    400


endpoint : String
endpoint =
    "https://sourcegraph.test:3443/.api"



-- MAIN


main : Program () Model Msg
main =
    Browser.element
        { init = init
        , update = update
        , view = view
        , subscriptions = subscriptions
        }



-- MODEL


type alias DataValue =
    { name : String
    , value : Float
    }


type alias Filter a =
    { a
        | dataPoints : Int
        , sortByCount : Bool
        , reverse : Bool
        , excludeStopWords : Bool
    }


type alias Model =
    { query : String
    , debounce : Int
    , dataPoints : Int
    , sortByCount : Bool
    , reverse : Bool
    , excludeStopWords : Bool
    , selectedTab : Tab
    , resultsMap : Dict String DataValue

    -- Debug client only
    , serverless : Bool
    }


init : () -> ( Model, Cmd Msg )
init _ =
    ( { query = "repo:.* content:output((.|\\n)* -> $date) type:commit count:all"
      , dataPoints = 30
      , sortByCount = True
      , reverse = False
      , excludeStopWords = False
      , selectedTab = Chart
      , debounce = 0
      , resultsMap = Dict.empty
      , serverless = False
      }
    , Task.perform identity (Task.succeed RunCompute)
    )



-- PORTS


type alias RawEvent =
    { data : String
    , eventType : Maybe String
    , id : Maybe String
    }


port receiveEvent : (RawEvent -> msg) -> Sub msg


port openStream : ( String, Maybe String ) -> Cmd msg



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.batch [ receiveEvent eventDecoder ]


eventDecoder : RawEvent -> Msg
eventDecoder event =
    case event.eventType of
        Just "results" ->
            OnResults (resultEventDecoder event.data)

        Just "done" ->
            ResultStreamDone

        _ ->
            NoOp


resultEventDecoder : String -> List Result
resultEventDecoder input =
    case Decode.decodeString (Decode.list resultDecoder) input of
        Ok results ->
            results

        Err _ ->
            []



-- UPDATE


type Msg
    = -- User inputs
      OnQueryChanged String
    | OnDebounce
    | OnDataPoints String
    | OnSortByCheckbox Bool
    | OnReverseCheckbox Bool
    | OnExcludeStopWordsCheckbox Bool
    | OnTabSelected Tab
      -- Data processing
    | RunCompute
    | OnResults (List Result)
    | ResultStreamDone
    | NoOp


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        OnQueryChanged newQuery ->
            ( { model | query = newQuery, debounce = model.debounce + 1 }
            , Task.perform (\_ -> OnDebounce) (Process.sleep debounceQueryInputMillis)
            )

        OnDebounce ->
            if model.debounce - 1 == 0 then
                update RunCompute { model | debounce = model.debounce - 1 }

            else
                ( { model | debounce = model.debounce - 1 }, Cmd.none )

        OnSortByCheckbox sortByCount ->
            ( { model | sortByCount = sortByCount }, Cmd.none )

        OnReverseCheckbox reverse ->
            ( { model | reverse = reverse }, Cmd.none )

        OnExcludeStopWordsCheckbox excludeStopWords ->
            ( { model | excludeStopWords = excludeStopWords }, Cmd.none )

        OnDataPoints i ->
            let
                newDataPoints =
                    case String.toInt i of
                        Just n ->
                            n

                        Nothing ->
                            0
            in
            ( { model | dataPoints = newDataPoints }, Cmd.none )

        OnTabSelected selectedTab ->
            ( { model | selectedTab = selectedTab }, Cmd.none )

        RunCompute ->
            if model.serverless then
                ( { model | resultsMap = exampleResultsMap }, Cmd.none )

            else
                ( { model | resultsMap = Dict.empty }
                , openStream ( endpoint ++ Url.Builder.absolute [ "compute", "stream" ] [ Url.Builder.string "q" model.query ], Nothing )
                )

        OnResults r ->
            ( { model | resultsMap = List.foldl updateResultsMap model.resultsMap (parseResults r) }
            , Cmd.none
            )

        ResultStreamDone ->
            ( model, Cmd.none )

        NoOp ->
            ( model, Cmd.none )



-- VIEW


table : List DataValue -> E.Element Msg
table data =
    let
        headerAttrs =
            [ F.bold
            , F.size 12
            , F.color darkModeFontColor
            , E.padding 5
            , Border.widthEach { bottom = 1, top = 0, left = 0, right = 0 }
            ]
    in
    E.el [ E.padding 100, E.centerX ]
        (E.table [ E.width E.fill ]
            { data = data
            , columns =
                [ { header = E.el headerAttrs (E.text " ")
                  , width = E.fillPortion 2
                  , view = \v -> E.el [ E.padding 5 ] (E.el [ E.width E.fill, E.padding 10, Border.rounded 5, Border.width 1 ] (E.text v.name))
                  }
                , { header = E.el (headerAttrs ++ [ F.alignRight ]) (E.text "Count")
                  , width = E.fillPortion 1
                  , view =
                        \v ->
                            E.el
                                [ E.centerY
                                , F.size 12
                                , F.color darkModeFontColor
                                , F.alignRight
                                , E.padding 5
                                ]
                                (E.text (String.fromFloat v.value))
                  }
                ]
            }
        )


histogram : List DataValue -> E.Element Msg
histogram data =
    E.el
        [ E.width E.fill
        , E.height (E.fill |> E.minimum 400)
        , E.centerX
        , E.alignTop
        , E.padding 30
        ]
        (E.html
            (C.chart
                [ CA.height 300, CA.width (toFloat width) ]
                [ C.bars
                    [ CA.spacing 0.0 ]
                    [ C.bar .value [ CA.color "#A112FF", CA.roundTop 0.2 ] ]
                    data
                , C.binLabels .name [ CA.moveDown 25, CA.rotate 45, CA.alignRight ]
                , C.barLabels [ CA.moveDown 12, CA.color "white", CA.fontSize 12 ]
                ]
            )
        )


dataView : List DataValue -> E.Element Msg
dataView data =
    E.row []
        [ E.el [ E.padding 10, E.alignLeft, E.width E.fill ]
            (E.column [] (List.map (\d -> E.text d.name) data))
        ]


inputRow : Model -> E.Element Msg
inputRow model =
    E.el [ E.centerX, E.width E.fill ]
        (E.column [ E.width E.fill ]
            [ I.text [ Background.color darkModeTextInputColor ]
                { onChange = OnQueryChanged
                , placeholder = Nothing
                , text = model.query
                , label = I.labelHidden ""
                }
            , E.row [ E.paddingXY 0 10 ]
                [ I.text [ E.width (E.fill |> E.maximum 65), F.center, Background.color darkModeTextInputColor ]
                    { onChange = OnDataPoints
                    , placeholder = Nothing
                    , text =
                        case model.dataPoints of
                            0 ->
                                ""

                            n ->
                                String.fromInt n
                    , label = I.labelHidden ""
                    }
                , I.checkbox [ E.paddingXY 10 0 ]
                    { onChange = OnSortByCheckbox
                    , icon = I.defaultCheckbox
                    , checked = model.sortByCount
                    , label = I.labelRight [] (E.text "sort by count")
                    }
                , I.checkbox [ E.paddingXY 10 0 ]
                    { onChange = OnReverseCheckbox
                    , icon = I.defaultCheckbox
                    , checked = model.reverse
                    , label = I.labelRight [] (E.text "reverse")
                    }
                , I.checkbox [ E.paddingXY 10 0 ]
                    { onChange = OnExcludeStopWordsCheckbox
                    , icon = I.defaultCheckbox
                    , checked = model.excludeStopWords
                    , label = I.labelRight [] (E.text "exclude stop words")
                    }
                ]
            ]
        )


type Tab
    = Chart
    | Table
    | Data


color =
    { skyBlue = E.rgb255 0x00 0xCB 0xEC
    , vividViolet = E.rgb255 0xA1 0x12 0xFF
    , vermillion = E.rgb255 0xFF 0x55 0x43
    }


tab : Tab -> Tab -> E.Element Msg
tab thisTab selectedTab =
    let
        isSelected =
            thisTab == selectedTab

        padOffset =
            if isSelected then
                0

            else
                2

        borderWidths =
            if isSelected then
                { left = 1, top = 1, right = 1, bottom = 0 }

            else
                { bottom = 1, top = 0, left = 0, right = 0 }

        corners =
            if isSelected then
                { topLeft = 6, topRight = 6, bottomLeft = 0, bottomRight = 0 }

            else
                { topLeft = 0, topRight = 0, bottomLeft = 0, bottomRight = 0 }

        tabColor =
            case selectedTab of
                Chart ->
                    color.vividViolet

                Table ->
                    color.vermillion

                Data ->
                    color.skyBlue

        text =
            case thisTab of
                Chart ->
                    "Chart"

                Table ->
                    "Table"

                Data ->
                    "Data"
    in
    E.el
        [ Border.widthEach borderWidths
        , Border.roundEach corners
        , Border.color tabColor
        , Element.Events.onClick (OnTabSelected thisTab)
        , E.htmlAttribute (Html.Attributes.style "cursor" "pointer")
        , E.width E.fill
        ]
        (E.el
            [ E.centerX
            , E.width E.fill
            , E.centerY
            , E.paddingEach { left = 30, right = 30, top = 10 + padOffset, bottom = 10 - padOffset }
            ]
            (E.text text)
        )


outputRow : Tab -> E.Element Msg
outputRow selectedTab =
    E.row [ E.centerX, E.width E.fill ]
        [ tab Chart selectedTab
        , tab Table selectedTab
        , tab Data selectedTab
        ]


view : Model -> Html Msg
view model =
    E.layout
        [ E.width E.fill
        , F.family [ F.typeface "Fira Code" ]
        , F.size 12
        , F.color darkModeFontColor
        , Background.color darkModeBackgroundColor
        ]
        (E.row [ E.centerX, E.width (E.fill |> E.maximum width) ]
            [ E.column [ E.centerX, E.width (E.fill |> E.maximum width), E.paddingXY 20 20 ]
                [ inputRow model
                , outputRow model.selectedTab
                , let
                    data =
                        Dict.toList model.resultsMap
                            |> List.map Tuple.second
                            |> filterData model
                  in
                  case model.selectedTab of
                    Chart ->
                        histogram data

                    Table ->
                        table data

                    Data ->
                        dataView data
                ]
            ]
        )



-- DATA LOGIC


parseResults : List Result -> List String
parseResults l =
    List.filterMap
        (\r ->
            case r of
                Output v ->
                    String.split "\n" v.value
                        |> List.filter (not << String.isEmpty)
                        |> Just

                ReplaceInPlace v ->
                    Just [ v.value ]
        )
        l
        |> List.concat


updateResultsMap : String -> Dict String DataValue -> Dict String DataValue
updateResultsMap textResult =
    Dict.update
        textResult
        (\v ->
            case v of
                Nothing ->
                    Just { name = textResult, value = 1 }

                Just existing ->
                    Just { existing | value = existing.value + 1 }
        )


filterData : Filter a -> List DataValue -> List DataValue
filterData { dataPoints, sortByCount, reverse, excludeStopWords } data =
    let
        pipeSort =
            if sortByCount then
                List.sortWith
                    (\l r ->
                        if l.value < r.value then
                            GT

                        else if l.value > r.value then
                            LT

                        else
                            EQ
                    )

            else
                identity
    in
    let
        pipeReverse =
            if reverse then
                List.reverse

            else
                identity
    in
    let
        pipeStopWords =
            if excludeStopWords then
                List.filter (\{ name } -> not (Dict.member (String.toLower name) Dict.empty))

            else
                identity
    in
    data
        |> pipeStopWords
        |> pipeSort
        |> pipeReverse
        |> List.take dataPoints
        |> pipeReverse



-- STREAMING RESULT TYPES


type Result
    = Output TextResult
    | ReplaceInPlace TextResult


type alias TextResult =
    { value : String
    , repository : Maybe String
    , commit : Maybe String
    , path : Maybe String
    }



-- DECODERS


resultDecoder : Decoder Result
resultDecoder =
    field "kind" Decode.string
        |> Decode.andThen
            (\t ->
                case t of
                    "replace-in-place" ->
                        textResultDecoder
                            |> Decode.map ReplaceInPlace

                    "output" ->
                        textResultDecoder
                            |> Decode.map Output

                    _ ->
                        fail ("Unrecognized type " ++ t)
            )


textResultDecoder : Decoder TextResult
textResultDecoder =
    Decode.succeed TextResult
        |> Json.Decode.Pipeline.required "value" Decode.string
        |> Json.Decode.Pipeline.optional "repository" (maybe Decode.string) Nothing
        |> Json.Decode.Pipeline.optional "commit" (maybe Decode.string) Nothing
        |> Json.Decode.Pipeline.optional "path" (maybe Decode.string) Nothing



-- STYLING


darkModeBackgroundColor : E.Color
darkModeBackgroundColor =
    E.rgb255 0x18 0x1B 0x26


darkModeFontColor : E.Color
darkModeFontColor =
    E.rgb255 0xFF 0xFF 0xFF


darkModeTextInputColor : E.Color
darkModeTextInputColor =
    E.rgb255 0x1D 0x22 0x2F



-- DEBUG DATA


exampleResultsMap : Dict String DataValue
exampleResultsMap =
    [ { name = "Errorf"
      , value = 10.0
      }
    , { name = "Func\nmulti\nline"
      , value = 5.0
      }
    , { name = "Qux"
      , value = 2.0
      }
    ]
        |> List.map (\d -> ( d.name, d ))
        |> Dict.fromList
