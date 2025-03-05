using System;
using System.Collections.Generic;

public static class SliceBuiltins
{
    public static List<T> Append<T>(this List<T> list, T element)
    {
        var result = new List<T>(list);
        result.Add(element);
        return result;
    }

    public static List<T> Append<T>(this List<T> list, params T[] elements)
    {
        var result = new List<T>(list);
        result.AddRange(elements);
        return result;
    }

    public static List<T> Append<T>(this List<T> list, List<T> elements)
    {
        var result = new List<T>(list);
        result.AddRange(elements);
        return result;
    }
}

