// hooks.js - Typed Redux hooks
// Provides type-safe hooks for accessing Redux state and dispatch

import { useDispatch, useSelector } from 'react-redux';

// Use throughout your app instead of plain `useDispatch` and `useSelector`
export const useAppDispatch = useDispatch;
export const useAppSelector = useSelector;

