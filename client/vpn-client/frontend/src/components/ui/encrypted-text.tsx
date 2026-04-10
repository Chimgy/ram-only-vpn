import React, { useEffect, useRef, useState } from "react";
import { motion, useInView } from "motion/react";
import { cn } from "@/lib/utils";

type EncryptedTextProps = {
  text: string;
  className?: string;
  revealDelayMs?: number;
  charset?: string;
  flipDelayMs?: number;
  encryptedClassName?: string;
  revealedClassName?: string;
  /** Increment to re-trigger animation (used when revealed is undefined) */
  trigger?: number;
  /**
   * When provided, directly controls state:
   *   false = continuously scramble (no reveal)
   *   true  = play reveal animation
   * When omitted, falls back to isInView (original behaviour).
   */
  revealed?: boolean;
};

const DEFAULT_CHARSET =
  "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+-={}[];:,.<>/?";

function generateRandomCharacter(charset: string): string {
  return charset.charAt(Math.floor(Math.random() * charset.length));
}

function generateGibberishPreservingSpaces(original: string, charset: string): string {
  let result = "";
  for (let i = 0; i < original.length; i++) {
    result += original[i] === " " ? " " : generateRandomCharacter(charset);
  }
  return result;
}

export const EncryptedText: React.FC<EncryptedTextProps> = ({
  text,
  className,
  revealDelayMs = 50,
  charset = DEFAULT_CHARSET,
  flipDelayMs = 50,
  encryptedClassName,
  revealedClassName,
  trigger,
  revealed,
}) => {
  const ref = useRef<HTMLSpanElement>(null);
  const isInView = useInView(ref, { once: true });

  const [revealCount, setRevealCount] = useState(0);
  const [, setScrambleTick] = useState(0); // forces re-render during scramble mode
  const animFrameRef = useRef<number | null>(null);
  const startTimeRef = useRef<number>(0);
  const lastFlipTimeRef = useRef<number>(0);
  const scrambleCharsRef = useRef<string[]>(
    text ? generateGibberishPreservingSpaces(text, charset).split("") : [],
  );

  const shouldReveal = revealed !== undefined ? revealed : isInView;

  useEffect(() => {
    if (!text) return;

    if (animFrameRef.current !== null) {
      cancelAnimationFrame(animFrameRef.current);
      animFrameRef.current = null;
    }

    if (!shouldReveal) {
      // Continuously scramble — never advance revealCount
      setRevealCount(0);
      scrambleCharsRef.current = generateGibberishPreservingSpaces(text, charset).split("");

      let lastFlip = performance.now();
      let cancelled = false;

      const scrambleLoop = (now: number) => {
        if (cancelled) return;
        if (now - lastFlip >= Math.max(0, flipDelayMs)) {
          scrambleCharsRef.current = generateGibberishPreservingSpaces(text, charset).split("");
          setScrambleTick(t => t + 1);
          lastFlip = now;
        }
        animFrameRef.current = requestAnimationFrame(scrambleLoop);
      };

      animFrameRef.current = requestAnimationFrame(scrambleLoop);
      return () => {
        cancelled = true;
        if (animFrameRef.current !== null) cancelAnimationFrame(animFrameRef.current);
      };
    }

    // Reveal animation
    scrambleCharsRef.current = generateGibberishPreservingSpaces(text, charset).split("");
    startTimeRef.current = performance.now();
    lastFlipTimeRef.current = startTimeRef.current;
    setRevealCount(0);

    let cancelled = false;

    const update = (now: number) => {
      if (cancelled) return;

      const elapsedMs = now - startTimeRef.current;
      const totalLength = text.length;
      const currentRevealCount = Math.min(totalLength, Math.floor(elapsedMs / Math.max(1, revealDelayMs)));

      setRevealCount(currentRevealCount);

      if (currentRevealCount >= totalLength) return;

      const timeSinceLastFlip = now - lastFlipTimeRef.current;
      if (timeSinceLastFlip >= Math.max(0, flipDelayMs)) {
        for (let i = currentRevealCount; i < totalLength; i++) {
          if (text[i] !== " ") scrambleCharsRef.current[i] = generateRandomCharacter(charset);
        }
        lastFlipTimeRef.current = now;
      }

      animFrameRef.current = requestAnimationFrame(update);
    };

    animFrameRef.current = requestAnimationFrame(update);
    return () => {
      cancelled = true;
      if (animFrameRef.current !== null) cancelAnimationFrame(animFrameRef.current);
    };
  }, [shouldReveal, text, revealDelayMs, charset, flipDelayMs, trigger]);

  if (!text) return null;

  return (
    <motion.span ref={ref} className={cn(className)} aria-label={text} role="text">
      {text.split("").map((char, index) => {
        const isRevealed = index < revealCount;
        const displayChar = isRevealed
          ? char
          : char === " "
            ? " "
            : (scrambleCharsRef.current[index] ?? generateRandomCharacter(charset));

        return (
          <span key={index} className={cn(isRevealed ? revealedClassName : encryptedClassName)}>
            {displayChar}
          </span>
        );
      })}
    </motion.span>
  );
};
