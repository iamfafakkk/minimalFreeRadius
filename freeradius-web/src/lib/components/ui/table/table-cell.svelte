<script lang="ts">
	import { cn } from '$lib/utils.js';
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<td class={cn('p-2 align-middle [&:has([role=checkbox])]:pr-0', klass)} {...$$restProps}>
	<slot />
</td>
