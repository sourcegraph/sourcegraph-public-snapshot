import React from "react";

export interface RangePosition {
  startOffset: number;
  endOffset: number;
  isExact: boolean;
}

export class HighlightedTextProps {
  constructor(readonly text: string, readonly positions: RangePosition[]) {}
  offsetSum(): number {
    let sum = 0;
    this.positions.forEach((pos) => {
      sum += pos.startOffset;
    });
    return sum;
  }
  exactCount(): number {
    let result = 0;
    this.positions.forEach((pos) => {
      if (pos.isExact) {
        result++;
      }
    });
    return result;
  }
  isExact(): boolean {
    return this.positions.length === 1 && this.positions[0].isExact;
  }
}

interface HighlightedTextPropsInstance {
  value: HighlightedTextProps;
}

export const HighlightedText: React.FunctionComponent<HighlightedTextPropsInstance> = (
  propsInstance
) => {
  const props = propsInstance.value;
  const spans: JSX.Element[] = [];
  let start = 0;
  function pushSpan(
    className: string,
    startOffset: number,
    endOffset: number
  ): void {
    if (startOffset >= endOffset) return;
    const text = props.text.substring(startOffset, endOffset);
    const key = `${text}-${className}`;
    const span = (
      <span key={key} className={className}>
        {text}
      </span>
    );
    spans.push(span);
  }
  for (let i = 0; i < props.positions.length; i++) {
    const pos = props.positions[i];
    if (pos.startOffset > start) {
      pushSpan("fuzzy-files-plaintext", start, pos.startOffset);
    }
    start = pos.endOffset;
    pushSpan("fuzzy-files-highlighted", pos.startOffset, pos.endOffset);
  }
  pushSpan("fuzzy-files-plaintext", start, props.text.length);

  return <>{spans}</>;
};
