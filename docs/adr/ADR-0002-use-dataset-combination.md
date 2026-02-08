# ADR-0002: Use Combination of Datasets

## Context
We have been using just RAVDESS which is very small and not very diverse.

## Decision
We will train the tone 3.0 model on a combination of several different datasets.

## Rationale
This will result in a more diverse training base, allowing the model to better generalize to noisy real-world situations.

## Alternatives Considered
Considered using a single sepereate dataset.

## Consequences
### Positive
- More accurate and robust model. 
  
## Negative
- More setup

## Follow-ups
- Create model
